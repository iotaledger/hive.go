package syncutils

import (
	"sync"

	"github.com/iotaledger/hive.go/core/generics/event"
)

type Counter struct {
	value         int
	mutex         sync.RWMutex
	updatedEvent  *event.Linkable[*counterEvent]
	increasedCond *sync.Cond
	decreasedCond *sync.Cond
}

func NewCounter() (newCounter *Counter) {
	newCounter = new(Counter)
	newCounter.updatedEvent = event.NewLinkable[*counterEvent]()
	newCounter.increasedCond = sync.NewCond(&newCounter.mutex)
	newCounter.decreasedCond = sync.NewCond(&newCounter.mutex)

	return
}

func (c *Counter) Subscribe(updateCallbacks ...func(oldValue, newValue int)) (unsubscribe func()) {
	if len(updateCallbacks) == 0 {
		return func() {}
	}

	closure := event.NewClosure(func(event *counterEvent) {
		for _, updateCallback := range updateCallbacks {
			updateCallback(event.oldValue, event.newValue)
		}
	})

	c.updatedEvent.Hook(closure)

	return func() {
		c.updatedEvent.Detach(closure)
	}
}

func (c *Counter) Get() (value int) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	return c.value
}

func (c *Counter) Set(newValue int) (oldValue int) {
	if oldValue = c.set(newValue); oldValue < newValue {
		c.increasedCond.Broadcast()
	} else if oldValue > newValue {
		c.decreasedCond.Broadcast()
	}

	return oldValue
}

func (c *Counter) Update(delta int) (newValue int) {
	if newValue = c.update(delta); delta > 1 {
		c.increasedCond.Broadcast()
	} else if delta < 1 {
		c.decreasedCond.Broadcast()
	}

	return newValue
}

func (c *Counter) Increase() (newValue int) {
	return c.Update(1)
}

func (c *Counter) Decrease() (newValue int) {
	return c.Update(-1)
}

func (c *Counter) WaitIsZero() {
	c.WaitIsBelow(1)
}

func (c *Counter) WaitIsBelow(threshold int) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	for c.value >= threshold {
		c.decreasedCond.Wait()
	}
}

func (c *Counter) WaitIsAbove(threshold int) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	for c.value <= threshold {
		c.increasedCond.Wait()
	}
}

func (c *Counter) set(newValue int) (oldValue int) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if oldValue = c.value; newValue != oldValue {
		c.updatedEvent.Trigger(&counterEvent{
			oldValue: oldValue,
			newValue: newValue,
		})
	}

	return oldValue
}

func (c *Counter) update(delta int) (newValue int) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if delta == 0 {
		return c.value
	}

	oldValue := c.value
	newValue = oldValue + delta

	c.updatedEvent.Trigger(&counterEvent{
		oldValue: oldValue,
		newValue: newValue,
	})

	return newValue
}

type counterEvent struct {
	oldValue int
	newValue int
}
