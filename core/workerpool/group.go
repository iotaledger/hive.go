package workerpool

import (
	"fmt"
	"strconv"
	"strings"
	"sync"

	"github.com/iotaledger/hive.go/core/generics/lo"
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
	root        *Group
}

func NewGroup(name string) (group *Group) {
	return newGroupWithRoot(name, nil)
}

func newGroupWithRoot(name string, root *Group) (group *Group) {
	return &Group{
		PendingChildrenCounter: syncutils.NewCounter(),
		name:                   name,
		pools:                  orderedmap.New[string, *UnboundedWorkerPool](),
		groups:                 orderedmap.New[string, *Group](),
		root:                   root,
	}
}

func (g *Group) Name() (name string) {
	return g.name
}

func (g *Group) CreatePool(name string, optsWorkerCount ...int) (pool *UnboundedWorkerPool) {
	g.poolsMutex.Lock()
	defer g.poolsMutex.Unlock()

	pool = NewUnboundedWorkerPool(name, optsWorkerCount...)
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

func (g *Group) Pool(name string) (pool *UnboundedWorkerPool, exists bool) {
	g.poolsMutex.RLock()
	defer g.poolsMutex.RUnlock()

	return g.pools.Get(name)
}

func (g *Group) Pools() (pools map[string]*UnboundedWorkerPool) {
	pools = make(map[string]*UnboundedWorkerPool)

	g.poolsMutex.RLock()
	g.pools.ForEach(func(name string, pool *UnboundedWorkerPool) bool {
		pools[fmt.Sprintf("%s.%s", g.name, name)] = pool
		return true
	})
	g.poolsMutex.RUnlock()

	g.groupsMutex.RLock()
	g.groups.ForEach(func(_ string, group *Group) bool {
		for name, pool := range group.Pools() {
			pools[fmt.Sprintf("%s.%s", g.name, name)] = pool
		}
		return true
	})
	g.groupsMutex.RUnlock()

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
	g.groupsMutex.RLock()
	defer g.groupsMutex.RUnlock()

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
	g.pools.ForEach(func(key string, value *UnboundedWorkerPool) bool {
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

const indentationString = "    "
