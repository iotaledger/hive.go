package event

import (
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/hive.go/runtime/workerpool"
)

func Benchmark(b *testing.B) {
	testEvent := New1[int]()
	testEvent.Hook(func(int) {})

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		testEvent.Trigger(i)
	}
}

func TestTrigger_PreTrigger(t *testing.T) {
	var triggerCount atomic.Uint64
	var hookCount atomic.Uint64
	var preTriggerCount atomic.Uint64

	testEvent := New1[int](WithPreTriggerFunc(func(i int) {
		preTriggerCount.Add(1)
	}))
	testEvent.Hook(func(int) {
		hookCount.Add(1)
	})

	testEvent.Hook(func(int) {
		hookCount.Add(1)
	})

	for i := 0; i < 10; i++ {
		triggerCount.Add(1)
		testEvent.Trigger(i)
	}

	require.Equal(t, uint64(10), triggerCount.Load())
	require.Equal(t, uint64(20), preTriggerCount.Load())
	require.Equal(t, uint64(20), hookCount.Load())
}

func TestTrigger_LinkTo_PreTrigger(t *testing.T) {
	var triggerCount atomic.Uint64
	var hook1Count atomic.Uint64
	var hook2Count atomic.Uint64
	var preTriggerCount atomic.Uint64

	testEvent := New1[int](WithPreTriggerFunc(func(i int) {
		preTriggerCount.Add(1)
	}))
	testEvent2 := New1[int]()
	testEvent.Hook(func(int) {
		hook1Count.Add(1)
	})

	testEvent2.Hook(func(int) {
		hook2Count.Add(1)
	})

	testEvent2.LinkTo(testEvent)

	for i := 0; i < 10; i++ {
		triggerCount.Add(1)
		testEvent.Trigger(i)
	}

	require.Equal(t, uint64(10), triggerCount.Load())
	require.Equal(t, uint64(20), preTriggerCount.Load())
	require.Equal(t, uint64(10), hook1Count.Load())
	require.Equal(t, uint64(10), hook2Count.Load())
}

func TestTriggerSettings_MaxTriggerCount(t *testing.T) {
	var triggerCount atomic.Uint64

	testEvent := New1[int](WithMaxTriggerCount(3))
	testEvent.Hook(func(int) {
		triggerCount.Add(1)
	})

	for i := 0; i < 10; i++ {
		go testEvent.Trigger(i)
	}

	require.Eventually(t, func() bool {
		return triggerCount.Load() == 3
	}, 1*time.Second, 10*time.Millisecond)

	time.Sleep(1 * time.Second)

	require.Equal(t, uint64(3), triggerCount.Load())
}

func TestTriggerSettings_Hook_MaxTriggerCount(t *testing.T) {
	var triggerCount atomic.Uint64

	testEvent := New1[int]()
	testEvent.Hook(func(int) {
		triggerCount.Add(1)
	}, WithMaxTriggerCount(3))

	for i := 0; i < 10; i++ {
		go testEvent.Trigger(i)
	}

	require.Eventually(t, func() bool {
		return triggerCount.Load() == 3
	}, 1*time.Second, 10*time.Millisecond)

	time.Sleep(1 * time.Second)

	require.Equal(t, uint64(3), triggerCount.Load())
}

func TestEvent1_Hook_WorkerPool(t *testing.T) {
	workerPool := workerpool.New(t.Name()).Start()

	var eventFired atomic.Bool

	testEvent := New1[int]()
	hook := testEvent.Hook(func(int) {
		time.Sleep(1 * time.Second)

		eventFired.Store(true)
	}, WithWorkerPool(workerPool))
	require.Equal(t, workerPool, hook.WorkerPool())

	require.False(t, testEvent.WasTriggered())
	require.False(t, hook.WasTriggered())
	testEvent.Trigger(0)
	require.True(t, testEvent.WasTriggered())
	require.Equal(t, 1, testEvent.TriggerCount())
	require.Equal(t, testEvent.MaxTriggerCount(), 0)
	require.False(t, testEvent.MaxTriggerCountReached())
	require.True(t, hook.WasTriggered())

	require.False(t, eventFired.Load())
	require.Eventually(t, eventFired.Load, 5*time.Second, 100*time.Millisecond)
	require.True(t, hook.WasTriggered())
}

func TestEvent1_WithoutWorkerPool(t *testing.T) {
	var eventFired atomic.Bool

	testEvent := New1[int](WithWorkerPool(nil))
	testEvent.Hook(func(int) {
		time.Sleep(1 * time.Second)

		eventFired.Store(true)
	})
	require.Equal(t, (*workerpool.WorkerPool)(nil), testEvent.WorkerPool())

	testEvent.Trigger(0)

	require.True(t, eventFired.Load())
}

func TestEvent1_Hook_WithoutWorkerPool(t *testing.T) {
	workerPool := workerpool.New(t.Name()).Start()

	var eventFired atomic.Bool

	testEvent := New1[int](WithWorkerPool(workerPool))
	hook := testEvent.Hook(func(int) {
		time.Sleep(1 * time.Second)

		eventFired.Store(true)
	}, WithWorkerPool(nil))
	require.Nil(t, hook.WorkerPool())

	testEvent.Trigger(0)

	require.True(t, eventFired.Load())
}

func TestLink(t *testing.T) {
	sourceEvents := NewEvents()

	eventTriggered := 0
	subEventTriggered := 0
	linkedEvents := NewEvents(sourceEvents)
	linkedEvents.Event.Hook(func(int) { eventTriggered++ })
	linkedEvents.SubEvents.Event.Hook(func(error) { subEventTriggered++ })

	sourceEvents.Event.Trigger(7)
	require.Equal(t, eventTriggered, 1)
	require.Equal(t, subEventTriggered, 0)

	sourceEvents.SubEvents.Event.Trigger(nil)
	require.Equal(t, eventTriggered, 1)
	require.Equal(t, subEventTriggered, 1)

	linkedEvents.LinkTo(nil)

	sourceEvents.Event.Trigger(7)
	sourceEvents.SubEvents.Event.Trigger(nil)
	require.Equal(t, eventTriggered, 1)
	require.Equal(t, subEventTriggered, 1)

	linkedEvents.LinkTo(sourceEvents)

	sourceEvents.Event.Trigger(7)
	sourceEvents.SubEvents.Event.Trigger(nil)
	require.Equal(t, eventTriggered, 2)
	require.Equal(t, subEventTriggered, 2)
}

type Events struct {
	Event     *Event1[int]
	SubEvents *SubEvents

	Group[Events, *Events]
}

var NewEvents = CreateGroupConstructor(func() *Events {
	return &Events{
		Event:     New1[int](),
		SubEvents: NewSubEvents(),
	}
})

type SubEvents struct {
	Event *Event1[error]

	Group[SubEvents, *SubEvents]
}

var NewSubEvents = CreateGroupConstructor(func() *SubEvents {
	return &SubEvents{
		Event: New1[error](),
	}
})
