package workerpool

import (
	"sync"

	"github.com/iotaledger/hive.go/core/generics/orderedmap"
	"github.com/iotaledger/hive.go/core/syncutils"
)

type Group struct {
	PendingChildrenCounter *syncutils.Counter

	name        string
	pools       *orderedmap.OrderedMap[string, *UnboundedWorkerPool]
	poolsMutex  sync.RWMutex
	groups      *orderedmap.OrderedMap[string, *Group]
	groupsMutex sync.RWMutex
}

func NewGroup(name string) (group *Group) {
	return &Group{
		PendingChildrenCounter: syncutils.NewCounter(),
		name:                   name,
		pools:                  orderedmap.New[string, *UnboundedWorkerPool](),
		groups:                 orderedmap.New[string, *Group](),
	}
}

func (g *Group) Name() (name string) {
	return g.name
}

func (g *Group) CreatePool(name string, optsWorkerCount ...int) (pool *UnboundedWorkerPool) {
	g.poolsMutex.Lock()
	defer g.poolsMutex.Unlock()

	pool = NewUnboundedWorkerPool(optsWorkerCount...)
	pool.PendingTasksCounter.Subscribe(func(oldValue, newValue int) {
		if oldValue == 0 {
			g.PendingChildrenCounter.Increase()
		} else if newValue == 0 {
			g.PendingChildrenCounter.Decrease()
		}
	})

	g.pools.Set(name, pool)

	return pool.Start()
}

func (g *Group) Pool(name string) (pool *UnboundedWorkerPool, exists bool) {
	g.poolsMutex.RLock()
	defer g.poolsMutex.RUnlock()

	return g.pools.Get(name)
}

func (g *Group) CreateGroup(name string) (group *Group) {
	group = NewGroup(name)
	group.PendingChildrenCounter.Subscribe(func(oldValue, newValue int) {
		if oldValue == 0 {
			g.PendingChildrenCounter.Increase()
		} else if newValue == 0 {
			g.PendingChildrenCounter.Decrease()
		}
	})

	g.groups.Set(name, group)

	return group
}

func (g *Group) Group(name string) (pool *Group, exists bool) {
	g.groupsMutex.RLock()
	defer g.groupsMutex.RUnlock()

	return g.groups.Get(name)
}

func (g *Group) Shutdown() {
	g.PendingChildrenCounter.WaitIsZero()

	g.shutdown()
}

func (g *Group) shutdown() {
	g.shutdownPools()
	g.shutdownGroups()
}

func (g *Group) shutdownPools() {
	g.poolsMutex.RLock()
	defer g.poolsMutex.RUnlock()

	g.pools.ForEach(func(_ string, pool *UnboundedWorkerPool) bool {
		pool.Shutdown(true)

		return true
	})
}

func (g *Group) shutdownGroups() {
	g.groupsMutex.RLock()
	defer g.groupsMutex.RUnlock()

	g.groups.ForEach(func(_ string, group *Group) bool {
		group.shutdown()

		return true
	})
}
