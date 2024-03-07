package queue

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewQueue(t *testing.T) {
	q := New[int](2)
	require.NotNil(t, q)
	assert.Equal(t, 0, q.Size())
	assert.Equal(t, 2, q.Capacity())
}

func TestQueueOfferPoll(t *testing.T) {
	q := New[int](2)
	require.NotNil(t, q)

	// offer element to queue
	{
		assert.True(t, q.Offer(1))
		assert.Equal(t, 1, q.Size())

		assert.True(t, q.Offer(2))
		assert.Equal(t, 2, q.Size())

		assert.False(t, q.Offer(3))
	}

	// Poll element from queue
	{
		polledValue, ok := q.Poll()
		assert.True(t, ok)
		assert.Equal(t, 1, polledValue)
		assert.Equal(t, 1, q.Size())

		polledValue, ok = q.Poll()
		assert.True(t, ok)
		assert.Equal(t, 2, polledValue)
		assert.Equal(t, 0, q.Size())

		polledValue, ok = q.Poll()
		assert.False(t, ok)
		assert.Zero(t, polledValue)
		assert.Equal(t, 0, q.Size())

		// Offer the empty queue again
		assert.True(t, q.Offer(3))
		assert.Equal(t, 1, q.Size())
	}

	// ForceOffer elements to the queue
	{
		removedElement, wasRemoved := q.ForceOffer(4)
		assert.False(t, wasRemoved)
		assert.Equal(t, 0, removedElement)
		assert.Equal(t, 2, q.Size())

		removedElement, wasRemoved = q.ForceOffer(5)
		assert.True(t, wasRemoved)
		assert.Equal(t, 3, removedElement)
		assert.Equal(t, 2, q.Size())

		removedElement, wasRemoved = q.ForceOffer(6)
		assert.True(t, wasRemoved)
		assert.Equal(t, 4, removedElement)

		polledValue, ok := q.Poll()
		assert.True(t, ok)
		assert.Equal(t, 5, polledValue)
		assert.Equal(t, 1, q.Size())
	}
}

func TestQueueOfferConcurrencySafe(t *testing.T) {
	q := New[int](100)
	require.NotNil(t, q)

	// let 10 workers fill the queue
	workers := 10
	var wg sync.WaitGroup
	wg.Add(workers)
	for i := 0; i < workers; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < 10; j++ {
				q.Offer(j)
			}
		}()
	}
	wg.Wait()

	// check that all the elements are offered
	assert.Equal(t, 100, q.Size())

	counter := make([]int, 10)
	for i := 0; i < 100; i++ {
		value, ok := q.Poll()
		assert.True(t, ok)
		counter[value]++
	}
	assert.Equal(t, 0, q.Size())

	// check that the insert numbers are correct
	for i := 0; i < 10; i++ {
		assert.Equal(t, 10, counter[i])
	}
}

func TestQueuePollConcurrencySafe(t *testing.T) {
	q := New[int](100)
	require.NotNil(t, q)

	for j := 0; j < 100; j++ {
		q.Offer(j)
	}

	// let 10 workers poll the queue
	workers := 10
	var wg sync.WaitGroup
	wg.Add(workers)
	for i := 0; i < workers; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < 10; j++ {
				_, ok := q.Poll()
				assert.True(t, ok)
			}
		}()
	}
	wg.Wait()

	// check that all the elements are polled
	assert.Equal(t, 0, q.Size())
}
