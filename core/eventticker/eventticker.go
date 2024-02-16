package eventticker

import (
	"sync"
	"time"

	"github.com/iotaledger/hive.go/core/index"
	"github.com/iotaledger/hive.go/core/memstorage"
	"github.com/iotaledger/hive.go/crypto"
	"github.com/iotaledger/hive.go/ds/shrinkingmap"
	"github.com/iotaledger/hive.go/runtime/options"
	"github.com/iotaledger/hive.go/runtime/timed"
)

// region EventTicker //////////////////////////////////////////////////////////////////////////////////////////////////

// EventTicker takes care of requesting blocks.
type EventTicker[I index.Type, T index.IndexedID[I]] struct {
	Events *Events[I, T]

	timedExecutor             *timed.Executor
	scheduledTickers          *memstorage.IndexedStorage[I, T, *timed.ScheduledTask]
	scheduledTickerCount      int
	scheduledTickerCountMutex sync.RWMutex
	lastEvictedIndex          I
	evictionMutex             sync.RWMutex

	optsRetryInterval       time.Duration
	optsRetryJitter         time.Duration
	optsMaxRequestThreshold int
}

// New creates a new block requester.
func New[I index.Type, T index.IndexedID[I]](opts ...options.Option[EventTicker[I, T]]) *EventTicker[I, T] {
	return options.Apply(&EventTicker[I, T]{
		Events: NewEvents[I, T](),

		timedExecutor:    timed.NewExecutor(1),
		scheduledTickers: memstorage.NewIndexedStorage[I, T, *timed.ScheduledTask](),

		optsRetryInterval:       10 * time.Second,
		optsRetryJitter:         5 * time.Second,
		optsMaxRequestThreshold: 100,
	}, opts)
}

func (r *EventTicker[I, T]) StartTickers(ids []T) {
	for _, id := range ids {
		r.StartTicker(id)
	}
}

func (r *EventTicker[I, T]) StartTicker(id T) {
	if r.addTickerToQueue(id) {
		r.Events.TickerStarted.Trigger(id)
		r.Events.Tick.Trigger(id)
	}
}

func (r *EventTicker[I, T]) StopTicker(id T) {
	if r.stopTicker(id) {
		r.Events.TickerStopped.Trigger(id)
	}
}

func (r *EventTicker[I, T]) HasTicker(id T) bool {
	r.evictionMutex.RLock()
	defer r.evictionMutex.RUnlock()

	if id.Index() <= r.lastEvictedIndex {
		return false
	}

	if queue := r.scheduledTickers.Get(id.Index(), false); queue != nil {
		return queue.Has(id)
	}

	return false
}

func (r *EventTicker[I, T]) QueueSize() int {
	r.scheduledTickerCountMutex.RLock()
	defer r.scheduledTickerCountMutex.RUnlock()

	return r.scheduledTickerCount
}

func (r *EventTicker[I, T]) EvictUntil(index I) {
	r.evictionMutex.Lock()
	defer r.evictionMutex.Unlock()

	if index <= r.lastEvictedIndex {
		return
	}

	for currentIndex := r.lastEvictedIndex + 1; currentIndex <= index; currentIndex++ {
		if evictedStorage := r.scheduledTickers.Evict(currentIndex); evictedStorage != nil {
			//nolint:revive // better be explicit here
			evictedStorage.ForEach(func(id T, scheduledTask *timed.ScheduledTask) bool {
				scheduledTask.Cancel()

				return true
			})

			r.updateScheduledTickerCount(-evictedStorage.Size())
		}
	}
	r.lastEvictedIndex = index
}

func (r *EventTicker[I, T]) Clear() {
	pendingTickers := make([]T, 0)
	//nolint:revive // better be explicit here
	r.scheduledTickers.ForEach(func(index I, storage *shrinkingmap.ShrinkingMap[T, *timed.ScheduledTask]) {
		storage.ForEach(func(id T, _ *timed.ScheduledTask) bool {
			pendingTickers = append(pendingTickers, id)
			return true
		})
	})

	for _, id := range pendingTickers {
		r.StopTicker(id)
	}
}

func (r *EventTicker[I, T]) Shutdown() {
	r.timedExecutor.Shutdown(timed.CancelPendingElements)
}

func (r *EventTicker[I, T]) addTickerToQueue(id T) (added bool) {
	r.evictionMutex.RLock()
	defer r.evictionMutex.RUnlock()

	if id.Index() <= r.lastEvictedIndex {
		return false
	}

	// ignore already scheduled requests
	queue := r.scheduledTickers.Get(id.Index(), true)
	if _, exists := queue.Get(id); exists {
		return false
	}

	// schedule the next request and trigger the event
	if scheduledTask := r.timedExecutor.ExecuteAfter(r.createReScheduler(id, 0), r.optsRetryInterval+time.Duration(crypto.Randomness.Float64()*float64(r.optsRetryJitter))); scheduledTask != nil {
		queue.Set(id, scheduledTask)
	}

	r.updateScheduledTickerCount(1)

	return true
}

func (r *EventTicker[I, T]) stopTicker(id T) (stopped bool) {
	r.evictionMutex.RLock()
	defer r.evictionMutex.RUnlock()

	storage := r.scheduledTickers.Get(id.Index())
	if storage == nil {
		return false
	}

	timer, exists := storage.Get(id)

	if !exists {
		return false
	}
	timer.Cancel()
	storage.Delete(id)

	r.updateScheduledTickerCount(-1)

	return true
}

func (r *EventTicker[I, T]) reSchedule(id T, count int) {
	r.Events.Tick.Trigger(id)

	// as we schedule a request at most once per id we do not need to make the trigger and the re-schedule atomic
	r.evictionMutex.RLock()
	defer r.evictionMutex.RUnlock()

	// reschedule, if the request has not been stopped in the meantime

	tickerStorage := r.scheduledTickers.Get(id.Index())
	if tickerStorage == nil {
		return
	}

	if _, requestExists := tickerStorage.Get(id); requestExists {
		// increase the request counter
		count++

		// if we have requested too often => stop the requests
		if count > r.optsMaxRequestThreshold {
			tickerStorage.Delete(id)

			r.updateScheduledTickerCount(-1)

			r.Events.TickerFailed.Trigger(id)

			return
		}

		if scheduledTask := r.timedExecutor.ExecuteAfter(r.createReScheduler(id, count), r.optsRetryInterval+time.Duration(crypto.Randomness.Float64()*float64(r.optsRetryJitter))); scheduledTask != nil {
			tickerStorage.Set(id, scheduledTask)
		}
	}
}

func (r *EventTicker[I, T]) createReScheduler(blkID T, count int) func() {
	return func() {
		r.reSchedule(blkID, count)
	}
}

func (r *EventTicker[I, T]) updateScheduledTickerCount(diff int) {
	r.scheduledTickerCountMutex.Lock()
	defer r.scheduledTickerCountMutex.Unlock()

	r.scheduledTickerCount += diff
}

// endregion ///////////////////////////////////////////////////////////////////////////////////////////////////////////

// region Options //////////////////////////////////////////////////////////////////////////////////////////////////////

// RetryInterval creates an option which sets the retry interval to the given value.
func RetryInterval[I index.Type, T index.IndexedID[I]](interval time.Duration) options.Option[EventTicker[I, T]] {
	return func(requester *EventTicker[I, T]) {
		requester.optsRetryInterval = interval
	}
}

// RetryJitter creates an option which sets the retry jitter to the given value.
func RetryJitter[I index.Type, T index.IndexedID[I]](retryJitter time.Duration) options.Option[EventTicker[I, T]] {
	return func(requester *EventTicker[I, T]) {
		requester.optsRetryJitter = retryJitter
	}
}

// MaxRequestThreshold creates an option which defines how often the EventTicker should try to request blocks before
// canceling the request.
func MaxRequestThreshold[I index.Type, T index.IndexedID[I]](maxRequestThreshold int) options.Option[EventTicker[I, T]] {
	return func(requester *EventTicker[I, T]) {
		requester.optsMaxRequestThreshold = maxRequestThreshold
	}
}

// endregion ///////////////////////////////////////////////////////////////////////////////////////////////////////////
