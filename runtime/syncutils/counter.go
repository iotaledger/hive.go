package syncutils

import (
	"sync"

	"github.com/iotaledger/hive.go/ds/orderedmap"
)

type Counter struct {
	value              int
	valueMutex         sync.RWMutex
	valueIncreasedCond *sync.Cond
	valueDecreasedCond *sync.Cond
	subscribers        *orderedmap.OrderedMap[uint64, func(oldValue, newValue int)]
	subscribersCounter uint64
	subscribersMutex   sync.RWMutex
}

func NewCounter() (newCounter *Counter) {
	newCounter = new(Counter)
	newCounter.valueIncreasedCond = sync.NewCond(&newCounter.valueMutex)
	newCounter.valueDecreasedCond = sync.NewCond(&newCounter.valueMutex)
	newCounter.subscribers = orderedmap.New[uint64, func(oldValue int, newValue int)]()

	return
}

func (c *Counter) Get() (value int) {
	c.valueMutex.RLock()
	defer c.valueMutex.RUnlock()

	return c.value
}

func (c *Counter) Set(newValue int) (oldValue int) {
	if oldValue = c.set(newValue); oldValue < newValue {
		c.valueIncreasedCond.Broadcast()
	} else if oldValue > newValue {
		c.valueDecreasedCond.Broadcast()
	}

	return oldValue
}

func (c *Counter) Update(delta int) (newValue int) {
	if newValue = c.update(delta); delta >= 1 {
		c.valueIncreasedCond.Broadcast()
	} else if delta <= -1 {
		c.valueDecreasedCond.Broadcast()
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
	c.valueMutex.Lock()
	defer c.valueMutex.Unlock()

	for c.value >= threshold {
		c.valueDecreasedCond.Wait()
	}
}

func (c *Counter) WaitIsAbove(threshold int) {
	c.valueMutex.Lock()
	defer c.valueMutex.Unlock()

	for c.value <= threshold {
		c.valueIncreasedCond.Wait()
	}
}

func (c *Counter) Subscribe(subscribers ...func(oldValue, newValue int)) (unsubscribe func()) {
	if len(subscribers) == 0 {
		return func() {}
	}

	subscriberID := c.subscribe(func(oldValue, newValue int) {
		for _, updateCallback := range subscribers {
			updateCallback(oldValue, newValue)
		}
	})

	return func() {
		c.unsubscribe(subscriberID)
	}
}

func (c *Counter) set(newValue int) (oldValue int) {
	c.valueMutex.Lock()
	defer c.valueMutex.Unlock()

	if oldValue = c.value; newValue != oldValue {
		c.value = newValue

		c.notifySubscribers(oldValue, newValue)
	}

	return oldValue
}

func (c *Counter) update(delta int) (newValue int) {
	c.valueMutex.Lock()
	defer c.valueMutex.Unlock()

	oldValue := c.value
	if newValue = oldValue + delta; newValue != oldValue {
		c.value = newValue

		c.notifySubscribers(oldValue, newValue)
	}

	return newValue
}

func (c *Counter) subscribe(callback func(oldValue, newValue int)) (subscriptionID uint64) {
	c.subscribersMutex.Lock()
	defer c.subscribersMutex.Unlock()

	c.subscribersCounter++
	c.subscribers.Set(c.subscribersCounter, callback)

	return c.subscribersCounter
}

func (c *Counter) unsubscribe(subscriptionID uint64) {
	c.subscribersMutex.Lock()
	defer c.subscribersMutex.Unlock()

	c.subscribers.Delete(subscriptionID)
}

func (c *Counter) notifySubscribers(oldValue, newValue int) {
	c.subscribersMutex.RLock()
	defer c.subscribersMutex.RUnlock()

	c.subscribers.ForEach(func(_ uint64, subscription func(oldValue, newValue int)) bool {
		subscription(oldValue, newValue)

		return true
	})
}
