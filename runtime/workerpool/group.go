package workerpool

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/iotaledger/hive.go/core/generics/lo"
	"github.com/iotaledger/hive.go/core/generics/orderedmap"
	"github.com/iotaledger/hive.go/core/syncutils"
)

// Group is a group of WorkerPools that can be managed as a whole.
type Group struct {
	PendingChildrenCounter *syncutils.Counter

	name   string
	pools  *orderedmap.OrderedMap[string, *WorkerPool]
	groups *orderedmap.OrderedMap[string, *Group]
	root   *Group
}

// NewGroup creates a new Group.
func NewGroup(name string) (group *Group) {
	return newGroupWithRoot(name, nil)
}

// newGroupWithRoot creates a new Group with a root Group.
func newGroupWithRoot(name string, root *Group) (group *Group) {
	return &Group{
		PendingChildrenCounter: syncutils.NewCounter(),
		name:                   name,
		pools:                  orderedmap.New[string, *WorkerPool](),
		groups:                 orderedmap.New[string, *Group](),
		root:                   root,
	}
}

func (g *Group) Name() (name string) {
	return g.name
}

func (g *Group) CreatePool(name string, optsWorkerCount ...int) (pool *WorkerPool) {
	pool = New(name, optsWorkerCount...)
	pool.PendingTasksCounter.Subscribe(func(oldValue, newValue int) {
		if oldValue == 0 {
			g.PendingChildrenCounter.Increase()
		} else if newValue == 0 {
			g.PendingChildrenCounter.Decrease()
		}
	})

	if !g.pools.Set(name, pool) {
		panic(fmt.Sprintf("pool '%s' already exists", name))
	}

	return pool.Start()
}

func (g *Group) Root() *Group {
	return lo.Cond(g.root != nil, g.root, g)
}

func (g *Group) Wait() {
	g.PendingChildrenCounter.WaitIsZero()
}

func (g *Group) WaitAll() {
	g.Root().Wait()
}

func (g *Group) Pool(name string) (pool *WorkerPool, exists bool) {
	return g.pools.Get(name)
}

func (g *Group) Pools() (pools map[string]*WorkerPool) {
	pools = make(map[string]*WorkerPool)
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

func (g *Group) CreateGroup(name string) (group *Group) {
	group = newGroupWithRoot(name, g.Root())
	group.PendingChildrenCounter.Subscribe(func(oldValue, newValue int) {
		if oldValue == 0 {
			g.PendingChildrenCounter.Increase()
		} else if newValue == 0 {
			g.PendingChildrenCounter.Decrease()
		}
	})

	if !g.groups.Set(name, group) {
		panic(fmt.Sprintf("group '%s' already exists", name))
	}

	return group
}

func (g *Group) Group(name string) (pool *Group, exists bool) {
	return g.groups.Get(name)
}

func (g *Group) Shutdown() {
	g.PendingChildrenCounter.WaitIsZero()

	g.shutdown()
}

func (g *Group) String() (humanReadable string) {
	if indentedString := g.indentedString(0); indentedString != "" {
		return strings.TrimRight(g.indentedString(0), "\r\n")
	}

	return "> " + g.name + " (0 pending children)"
}

func (g *Group) indentedString(indentation int) (humanReadable string) {
	if pendingChildren := g.PendingChildrenCounter.Get(); pendingChildren != 0 {
		if children := g.childrenString(indentation + 1); children != "" {
			humanReadable = strings.Repeat(indentationString, indentation) + "> " + g.name + " (" + strconv.Itoa(pendingChildren) + " pending children) {\n"
			humanReadable += children
			humanReadable += strings.Repeat(indentationString, indentation) + "}\n"
		}
	}

	return humanReadable
}

func (g *Group) childrenString(indentation int) (humanReadable string) {
	humanReadable = g.poolsString(indentation)

	groups := g.groupsString(indentation)
	if humanReadable != "" && groups != "" {
		humanReadable += strings.Repeat(indentationString, indentation) + "\n"
	}

	return humanReadable + groups
}

func (g *Group) poolsString(indentation int) (humanReadable string) {
	g.pools.ForEach(func(key string, value *WorkerPool) bool {
		if currentValue := value.PendingTasksCounter.Get(); currentValue > 0 {
			humanReadable += strings.Repeat(indentationString, indentation) + "- " + key + " (" + strconv.Itoa(currentValue) + " pending tasks)\n"
		}

		return true
	})

	return humanReadable
}

func (g *Group) groupsString(indentation int) (humanReadable string) {
	g.groups.ForEach(func(key string, value *Group) bool {
		humanReadable += value.indentedString(indentation)
		return true
	})

	return humanReadable
}

func (g *Group) shutdown() {
	g.pools.ForEach(func(_ string, pool *WorkerPool) bool {
		pool.Shutdown(true)
		return true
	})

	g.groups.ForEach(func(_ string, group *Group) bool {
		group.shutdown()
		return true
	})
}

const indentationString = "    "
