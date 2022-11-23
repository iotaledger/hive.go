package syncutils

import (
	"sync"
)

type Counter struct {
	value          int
	mutex          sync.RWMutex
	valueIncreased *sync.Cond
	valueDecreased *sync.Cond
}

func NewCounter() (newCounter *Counter) {
	newCounter = new(Counter)
	newCounter.valueIncreased = sync.NewCond(&newCounter.mutex)
	newCounter.valueDecreased = sync.NewCond(&newCounter.mutex)

	return
}
func (b *Counter) Value() int {
	b.mutex.RLock()
	defer b.mutex.RUnlock()

	return b.value
}

func (b *Counter) Increase() {
	b.mutex.Lock()
	b.value++
	b.mutex.Unlock()

	b.valueIncreased.Broadcast()
}

func (b *Counter) Decrease() {
	b.mutex.Lock()
	b.value--
	b.mutex.Unlock()

	b.valueDecreased.Broadcast()
}

func (b *Counter) WaitIsZero() {
	b.WaitIsBelow(1)
}

func (b *Counter) WaitIsBelow(threshold int) {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	for b.value >= threshold {
		b.valueDecreased.Wait()
	}
}

func (b *Counter) WaitIsAbove(threshold int) {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	for b.value <= threshold {
		b.valueIncreased.Wait()
	}
}
