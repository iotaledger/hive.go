package causalordersync

import (
	"sync"

	"github.com/pkg/errors"

	"github.com/iotaledger/hive.go/core/index"
	"github.com/iotaledger/hive.go/core/memstorage"
	"github.com/iotaledger/hive.go/runtime/options"
	"github.com/iotaledger/hive.go/runtime/syncutils"
	"github.com/iotaledger/hive.go/runtime/workerpool"
)

// region CausalOrder ////////////////////////////////////////////////////////////////////////////////////////////////

// CausalOrder represents an order where an Entity is ordered after its causal dependencies (parents) have been ordered.
type CausalOrder[I index.Type, ID index.IndexedID[I], Entity OrderedEntity[I, ID]] struct {
	// entityProvider contains a function that provides the Entity that belongs to a given ID.
	entityProvider func(id ID) (entity Entity, exists bool)

	// isOrdered contains a function that determines if an Entity has been ordered already.
	isOrdered func(entity Entity) (isOrdered bool)

	// orderedCallback contains a function that is called when an Entity is ordered.
	orderedCallback func(entity Entity) (err error)

	// evictionCallback contains a function that is called whenever an Entity is evicted from the CausalOrderer.
	evictionCallback func(entity Entity, reason error)

	// parentsCallback contains a function that returns the parents of an Entity.
	parentsCallback func(entity Entity) []ID

	// checkReference contains a function that checks if a reference between a child and its parents is valid.
	checkReference func(child Entity, parent Entity) (err error)

	// unorderedParentsCounter contains an in-memory storage that keeps track of the unordered parents of an Entity.
	unorderedParentsCounter *memstorage.IndexedStorage[I, ID, uint8]

	// unorderedParentsCounterMutex contains a mutex used to synchronize access to the unorderedParentsCounter.
	unorderedParentsCounterMutex sync.Mutex

	// unorderedChildren contains an in-memory storage of the pending children of an unordered Entity.
	unorderedChildren *memstorage.IndexedStorage[I, ID, []Entity]

	// unorderedChildrenMutex contains a mutex used to synchronize access to the unorderedChildren.
	unorderedChildrenMutex sync.Mutex

	// lastEvictedIndex contains the last evicted slot.
	lastEvictedIndex I

	// evictionMutex contains the local manager used to orchestrate the eviction of old Entities.
	evictionMutex sync.RWMutex

	// dagMutex contains a mutex used to synchronize access to Entities.
	dagMutex *syncutils.DAGMutex[ID]

	workerPool *workerpool.WorkerPool
}

// New returns a new CausalOrderer instance with the given parameters.
func New[I index.Type, ID index.IndexedID[I], Entity OrderedEntity[I, ID]](
	workerPool *workerpool.WorkerPool,
	entityProvider func(id ID) (entity Entity, exists bool),
	isOrdered func(entity Entity) (isOrdered bool),
	orderedCallback func(entity Entity) (err error),
	evictionCallback func(entity Entity, reason error),
	parentsCallback func(entity Entity) []ID,
	opts ...options.Option[CausalOrder[I, ID, Entity]],
) (newCausalOrder *CausalOrder[I, ID, Entity]) {
	return options.Apply(&CausalOrder[I, ID, Entity]{
		workerPool:              workerPool,
		entityProvider:          entityProvider,
		isOrdered:               isOrdered,
		orderedCallback:         orderedCallback,
		evictionCallback:        evictionCallback,
		parentsCallback:         parentsCallback,
		checkReference:          checkReference[I, ID, Entity],
		unorderedParentsCounter: memstorage.NewIndexedStorage[I, ID, uint8](),
		unorderedChildren:       memstorage.NewIndexedStorage[I, ID, []Entity](),
		dagMutex:                syncutils.NewDAGMutex[ID](),
	}, opts)
}

// Queue adds the given Entity to the CausalOrderer and triggers it when it's ready.
func (c *CausalOrder[I, ID, Entity]) Queue(entity Entity) {
	for _, childToCheck := range c.triggerOrderedIfReady(entity) {
		c.triggerChildIfReady(childToCheck)
	}
}

// EvictUntil removes all Entities that are older than the given slot from the CausalOrder.
func (c *CausalOrder[I, ID, Entity]) EvictUntil(index I) {
	for _, evictedEntity := range c.evictUntil(index) {
		c.dagMutex.Lock(evictedEntity.ID())
		defer c.dagMutex.Unlock(evictedEntity.ID())

		c.evictionCallback(evictedEntity, errors.Errorf("entity evicted from %d", index))
	}
}

// triggerOrderedIfReady triggers the ordered callback of the given Entity if it's ready.
func (c *CausalOrder[I, ID, Entity]) triggerOrderedIfReady(entity Entity) []Entity {
	parents := c.parentsCallback(entity)

	c.evictionMutex.RLock()
	defer c.evictionMutex.RUnlock()
	c.dagMutex.RLock(parents...)
	defer c.dagMutex.RUnlock(parents...)
	c.dagMutex.Lock(entity.ID())
	defer c.dagMutex.Unlock(entity.ID())

	if c.isOrdered(entity) {
		return nil
	}

	if c.lastEvictedIndex >= entity.ID().Index() {
		c.evictionCallback(entity, errors.Errorf("entity %s below max evicted slot", entity.ID()))

		return nil
	}

	if !c.allParentsOrdered(entity) {
		return nil
	}

	return c.triggerOrderedCallback(entity)

}

// allParentsOrdered returns true if all parents of the given Entity are ordered.
func (c *CausalOrder[I, ID, Entity]) allParentsOrdered(entity Entity) (allParentsOrdered bool) {
	pendingParents := uint8(0)
	for _, parentID := range c.parentsCallback(entity) {
		parentEntity, exists := c.entityProvider(parentID)
		if !exists {
			c.evictionCallback(entity, errors.Errorf("parent %s not found", parentID))

			return
		}

		if err := c.checkReference(entity, parentEntity); err != nil {
			c.evictionCallback(entity, err)

			return
		}

		if !c.isOrdered(parentEntity) {
			pendingParents++

			c.registerUnorderedChild(parentID, entity)
		}
	}

	if pendingParents != 0 {
		c.setUnorderedParentsCounter(entity.ID(), pendingParents)
	}

	return pendingParents == 0
}

// registerUnorderedChild registers the given Entity as a child of the given parent ID.
func (c *CausalOrder[I, ID, Entity]) registerUnorderedChild(entityID ID, child Entity) {
	c.unorderedChildrenMutex.Lock()
	defer c.unorderedChildrenMutex.Unlock()

	unorderedChildrenStorage := c.unorderedChildren.Get(entityID.Index(), true)
	entityChildren, _ := unorderedChildrenStorage.Get(entityID)
	unorderedChildrenStorage.Set(entityID, append(entityChildren, child))
}

// setUnorderedParentsCounter sets the unordered parents counter of the given Entity to the given value.
func (c *CausalOrder[I, ID, Entity]) setUnorderedParentsCounter(entityID ID, unorderedParentsCount uint8) {
	c.unorderedParentsCounterMutex.Lock()
	defer c.unorderedParentsCounterMutex.Unlock()

	c.unorderedParentsCounter.Get(entityID.Index(), true).Set(entityID, unorderedParentsCount)
}

// decrementUnorderedParentsCounter decrements the unordered parents counter of the given Entity by 1 and returns the
// new value.
func (c *CausalOrder[I, ID, Entity]) decreaseUnorderedParentsCounter(metadata Entity) (newUnorderedParentsCounter uint8) {
	c.unorderedParentsCounterMutex.Lock()
	defer c.unorderedParentsCounterMutex.Unlock()

	unorderedParentsCounterStorage := c.unorderedParentsCounter.Get(metadata.ID().Index())
	newUnorderedParentsCounter, _ = unorderedParentsCounterStorage.Get(metadata.ID())
	newUnorderedParentsCounter--
	if newUnorderedParentsCounter == 0 {
		unorderedParentsCounterStorage.Delete(metadata.ID())

		return
	}

	unorderedParentsCounterStorage.Set(metadata.ID(), newUnorderedParentsCounter)

	return
}

// popUnorderedChild pops the children of the given parent ID from the unordered children storage.
func (c *CausalOrder[I, ID, Entity]) popUnorderedChildren(entityID ID) (pendingChildren []Entity) {
	c.unorderedChildrenMutex.Lock()
	defer c.unorderedChildrenMutex.Unlock()

	pendingChildrenStorage := c.unorderedChildren.Get(entityID.Index())
	if pendingChildrenStorage == nil {
		return pendingChildren
	}

	pendingChildren, _ = pendingChildrenStorage.Get(entityID)

	pendingChildrenStorage.Delete(entityID)

	return pendingChildren
}

// triggerChildIfReady triggers the ordered callback of the given Entity if it's unorderedParentsCounter reaches 0
// (after decreasing it).
func (c *CausalOrder[I, ID, Entity]) triggerChildIfReady(child Entity) {
	var childrenToCheck []Entity
	c.dagMutex.Lock(child.ID())
	if !c.isOrdered(child) && c.decreaseUnorderedParentsCounter(child) == 0 {
		childrenToCheck = c.triggerOrderedCallback(child)
	}
	c.dagMutex.Unlock(child.ID())

	for _, childToCheck := range childrenToCheck {
		c.triggerChildIfReady(childToCheck)
	}
}

// triggerOrderedCallback triggers the ordered callback of the given Entity and propagates .
func (c *CausalOrder[I, ID, Entity]) triggerOrderedCallback(entity Entity) []Entity {
	if err := c.orderedCallback(entity); err != nil {
		c.evictionCallback(entity, err)

		return nil
	}

	return c.popUnorderedChildren(entity.ID())
}

// entity returns the Entity with the given ID.
func (c *CausalOrder[I, ID, Entity]) entity(blockID ID) (entity Entity) {
	entity, exists := c.entityProvider(blockID)
	if !exists {
		panic(errors.Errorf("block %s does not exist", blockID))
	}

	return entity
}

// evictUntil evicts the given slot from the CausalOrder and returns the evicted Entities.
func (c *CausalOrder[I, ID, Entity]) evictUntil(index I) (evictedEntities map[ID]Entity) {
	c.evictionMutex.Lock()
	defer c.evictionMutex.Unlock()

	if index <= c.lastEvictedIndex {
		return
	}

	evictedEntities = make(map[ID]Entity)
	for currentIndex := c.lastEvictedIndex + 1; currentIndex <= index; currentIndex++ {
		c.evictEntitiesFromSlot(currentIndex, func(id ID) {
			if _, exists := evictedEntities[id]; !exists {
				evictedEntities[id] = c.entity(id)
			}
		})
	}
	c.lastEvictedIndex = index

	return evictedEntities
}

// evictEntitiesFromSlot evicts the Entities that belong to the given slot from the CausalOrder.
func (c *CausalOrder[I, ID, Entity]) evictEntitiesFromSlot(index I, entityCallback func(id ID)) {
	if childrenStorage := c.unorderedChildren.Get(index); childrenStorage != nil {
		childrenStorage.ForEachKey(func(id ID) bool {
			entityCallback(id)

			return true
		})
		c.unorderedChildren.Evict(index)
	}

	if unorderedParentsCountStorage := c.unorderedParentsCounter.Get(index); unorderedParentsCountStorage != nil {
		unorderedParentsCountStorage.ForEachKey(func(id ID) bool {
			entityCallback(id)

			return true
		})
		c.unorderedParentsCounter.Evict(index)
	}
}

// checkReference is the default function that checks if the given reference is valid.
func checkReference[I index.Type, ID index.IndexedID[I], Entity OrderedEntity[I, ID]](entity Entity, parent Entity) (err error) {
	return
}

// endregion ///////////////////////////////////////////////////////////////////////////////////////////////////////////

// region Options //////////////////////////////////////////////////////////////////////////////////////////////////////

// WithReferenceValidator is an option that sets the ReferenceValidator of the CausalOrder.
func WithReferenceValidator[I index.Type, ID index.IndexedID[I], Entity OrderedEntity[I, ID]](referenceValidator func(entity Entity, parent Entity) (err error)) options.Option[CausalOrder[I, ID, Entity]] {
	return func(causalOrder *CausalOrder[I, ID, Entity]) {
		causalOrder.checkReference = referenceValidator
	}
}

// endregion ///////////////////////////////////////////////////////////////////////////////////////////////////////////

// region Types ////////////////////////////////////////////////////////////////////////////////////////////////////////

// OrderedEntity is an interface that represents an Entity that can be causally ordered.
type OrderedEntity[I index.Type, ID index.IndexedID[I]] interface {
	// ID returns the ID of the Entity.
	ID() ID

	// comparable embeds the comparable interface.
	comparable
}

// endregion ///////////////////////////////////////////////////////////////////////////////////////////////////////////
