package workerpool

import (
	"fmt"
	"strconv"
	"strings"
	"sync/atomic"

	"github.com/iotaledger/hive.go/ds/orderedmap"
	"github.com/iotaledger/hive.go/lo"
	"github.com/iotaledger/hive.go/runtime/options"
	"github.com/iotaledger/hive.go/runtime/syncutils"
)

// Group is a group of WorkerPools that can be managed as a whole.
type Group struct {
	// PendingChildrenCounter contains the number of children that are currently pending.
	PendingChildrenCounter *syncutils.Counter

	// name is the name of the Group.
	name string

	// pools is a map of WorkerPools that are managed by this Group.
	pools *orderedmap.OrderedMap[string, *WorkerPool]

	// groups is a map of subgroups that are managed by this Group.
	groups *orderedmap.OrderedMap[string, *Group]

	// root is the root Group of this Group.
	root *Group

	// isShutdown is true if the group was shutdown.
	isShutdown atomic.Bool
}

// NewGroup creates a new Group.
func NewGroup(name string) *Group {
	return &Group{
		PendingChildrenCounter: syncutils.NewCounter(),
		name:                   name,
		pools:                  orderedmap.New[string, *WorkerPool](),
		groups:                 orderedmap.New[string, *Group](),
	}
}

// Name returns the name of the Group.
func (g *Group) Name() (name string) {
	return g.name
}

// CreatePool creates a new WorkerPool with the given name and returns it.
func (g *Group) CreatePool(name string, opts ...options.Option[WorkerPool]) *WorkerPool {
	workerPoolOpts := []options.Option[WorkerPool]{
		WithCancelPendingTasksOnShutdown(true),
	}
	workerPoolOpts = append(workerPoolOpts, opts...)

	pool := New(name, workerPoolOpts...)
	pool.PendingTasksCounter.Subscribe(func(oldValue, newValue int) {
		if oldValue == 0 {
			g.PendingChildrenCounter.Increase()
		} else if newValue == 0 {
			g.PendingChildrenCounter.Decrease()
		}
	})

	if previousPool, previousPoolExists := g.pools.Set(name, pool); previousPoolExists && previousPool.IsRunning() {
		panic(fmt.Sprintf("running pool '%s' already exists", name))
	}

	return pool.Start()
}

// Pool returns the WorkerPool with the given name.
func (g *Group) Pool(name string) (pool *WorkerPool, exists bool) {
	return g.pools.Get(name)
}

// Pools returns all WorkerPools of the Group.
func (g *Group) Pools() map[string]*WorkerPool {
	pools := make(map[string]*WorkerPool)

	g.pools.ForEach(func(name string, pool *WorkerPool) bool {
		pools[fmt.Sprintf("%s.%s", g.name, name)] = pool
		return true
	})

	g.groups.ForEach(func(_ string, group *Group) bool {
		for name, pool := range group.Pools() {
			pools[fmt.Sprintf("%s.%s", g.name, name)] = pool
		}

		return true
	})

	return pools
}

// CreateGroup creates a new Group with the given name and returns it.
func (g *Group) CreateGroup(name string) *Group {
	group := NewGroup(name)
	group.root = g.Root()
	group.PendingChildrenCounter.Subscribe(func(oldValue, newValue int) {
		if oldValue == 0 {
			g.PendingChildrenCounter.Increase()
		} else if newValue == 0 {
			g.PendingChildrenCounter.Decrease()
		}
	})

	if previousGroup, previousGroupExisted := g.groups.Set(name, group); previousGroupExisted && !previousGroup.IsShutdown() {
		panic(fmt.Sprintf("running group '%s' already exists", name))
	}

	return group
}

// Group returns the Group with the given name.
func (g *Group) Group(name string) (pool *Group, exists bool) {
	return g.groups.Get(name)
}

// WaitChildren waits until all children of the Group are idle.
func (g *Group) WaitChildren() {
	g.PendingChildrenCounter.WaitIsZero()
}

// WaitParents waits until all parents of the Group are idle.
func (g *Group) WaitParents() {
	g.Root().WaitChildren()
}

// Root returns the root Group of the Group.
func (g *Group) Root() *Group {
	return lo.Cond(g.root != nil, g.root, g)
}

// IsShutdown returns true if the group was shutdown.
func (g *Group) IsShutdown() bool {
	return g.isShutdown.Load()
}

// Shutdown shuts down all child elements of the Group.
func (g *Group) Shutdown() {
	g.PendingChildrenCounter.WaitIsZero()

	g.shutdown()
}

// String returns a human-readable string representation of the Group.
func (g *Group) String() string {
	if indentedString := g.string(0); indentedString != "" {
		return strings.TrimRight(g.string(0), "\r\n")
	}

	return "> " + g.name + " (0 pending children)"
}

// string returns a human-readable representation of the Group with the given indentation.
func (g *Group) string(indent int) string {
	pendingChildCounter := g.PendingChildrenCounter.Get()
	if pendingChildCounter == 0 {
		return ""
	}

	children := g.childrenString(indent + 1)
	if children == "" {
		return ""
	}

	result := strings.Repeat(indentStr, indent) + "> " + g.name + " (" + strconv.Itoa(pendingChildCounter) + " pending children) {\n"
	result += children
	result += strings.Repeat(indentStr, indent) + "}\n"

	return result
}

// childrenString returns a human-readable representation of the children of the Group with the given indentation.
func (g *Group) childrenString(indent int) string {
	pools := g.poolsString(indent)
	groups := g.groupsString(indent)

	if groups == "" {
		return pools
	}
	if pools == "" {
		return groups
	}

	return pools + strings.Repeat(indentStr, indent) + "\n" + groups
}

// poolsString returns a human-readable representation of the WorkerPools of the Group with the given indentation.
func (g *Group) poolsString(indent int) string {
	result := ""
	g.pools.ForEach(func(key string, value *WorkerPool) bool {
		if currentValue := value.PendingTasksCounter.Get(); currentValue > 0 {
			result += strings.Repeat(indentStr, indent) + "- " + key + " (" + strconv.Itoa(currentValue) + " pending tasks)\n"
		}

		return true
	})

	return result
}

// groupsString returns a human-readable representation of the Groups of the Group with the given indentation.
func (g *Group) groupsString(indent int) string {
	result := ""
	//nolint:revive // better be explicit here
	g.groups.ForEach(func(key string, value *Group) bool {
		result += value.string(indent)
		return true
	})

	return result
}

// shutdown shuts down all child elements of the Group.
func (g *Group) shutdown() {
	if g.isShutdown.Swap(true) {
		return
	}

	g.pools.ForEach(func(_ string, pool *WorkerPool) bool {
		pool.Shutdown()
		return true
	})

	g.groups.ForEach(func(_ string, group *Group) bool {
		group.shutdown()
		return true
	})
}

// indentStr is the string used for indentation.
const indentStr = "    "
