package eventticker_test

import (
	"encoding/binary"
	"math/rand"
	"sync"
	"testing"
	"time"

	"github.com/mr-tron/base58"
	"github.com/stretchr/testify/require"

	"github.com/iotaledger/hive.go/core/eventticker"
	"github.com/iotaledger/hive.go/lo"
)

func TestEventTicker(t *testing.T) {
	eventTicker := eventticker.New(
		eventticker.RetryInterval[index, testID](time.Second),
		eventticker.RetryJitter[index, testID](time.Millisecond),
	)

	// Add ticker for slot 1
	id := newtestID(index(1), randByte())
	eventTicker.StartTicker(id)

	has := eventTicker.HasTicker(id)
	require.True(t, has, "EventTicker.HasTicker() returned false, expected true")
	require.Equalf(t, 1, eventTicker.QueueSize(), "EventTicker.QueueSize() returned %d, expected 1", eventTicker.QueueSize())

	eventTicker.StartTicker(id)
	require.Equal(t, 1, eventTicker.QueueSize())

	// Stop and delete ticker for slot 1
	eventTicker.StopTicker(id)
	has = eventTicker.HasTicker(id)
	require.False(t, has, "EventTicker.HasTicker() returned true, expected false")
	require.Equalf(t, 0, eventTicker.QueueSize(), "EventTicker.QueueSize() returned %d, expected 0", eventTicker.QueueSize())

	// Add new tickers for slot 1 and 2
	id1 := newtestID(index(1), randByte())
	id2 := newtestID(index(2), randByte())
	eventTicker.StartTickers([]testID{id1, id2})

	require.Equalf(t, 2, eventTicker.QueueSize(), "EventTicker.QueueSize() returned %d, expected 2", eventTicker.QueueSize())
	has = eventTicker.HasTicker(id1)
	require.True(t, has, "EventTicker.HasTicker() returned false, expected true")
	has = eventTicker.HasTicker(id2)
	require.True(t, has, "EventTicker.HasTicker() returned false, expected true")

	// Evict slot 1
	eventTicker.EvictUntil(index(1))
	require.Equalf(t, 1, eventTicker.QueueSize(), "EventTicker.QueueSize() returned %d, expected 1", eventTicker.QueueSize())

	// Add ticker for evicted slot
	eventTicker.StartTicker(id1)
	has = eventTicker.HasTicker(id1)
	require.False(t, has, "EventTicker.HasTicker() returned true, expected false")

	// evict evicted slot
	eventTicker.EvictUntil(index(1))
	require.Equalf(t, 1, eventTicker.QueueSize(), "EventTicker.QueueSize() returned %d, expected 1", eventTicker.QueueSize())

	done := make(chan struct{})
	go func() {
		eventTicker.Shutdown()
		close(done)
	}()

	select {
	case <-done:
		// done
	case <-time.After(time.Second):
		t.Error("Shutdown timed out")
	}
}

func TestRescheduleTicker(t *testing.T) {
	eventTicker := eventticker.New(
		eventticker.RetryInterval[index, testID](time.Duration(time.Second)),
		eventticker.RetryJitter[index, testID](time.Duration(time.Millisecond)),
		eventticker.MaxRequestThreshold[index, testID](3),
	)
	var mutex sync.Mutex
	var startTime time.Time
	counter := 0
	done := make(chan struct{})
	id := newtestID(index(1), randByte())

	unhook := lo.Batch(
		eventTicker.Events.TickerStarted.Hook(func(ti testID) {
			mutex.Lock()
			defer mutex.Unlock()

			startTime = time.Now()
		}).Unhook,
		eventTicker.Events.Tick.Hook(func(ti testID) {
			mutex.Lock()
			defer mutex.Unlock()

			if counter > 0 {
				require.LessOrEqual(t, time.Since(startTime), time.Duration(1*time.Second+5*time.Millisecond))
				startTime = time.Now()
			}
			counter++
		}).Unhook,
		eventTicker.Events.TickerFailed.Hook(func(ti testID) {
			require.Equal(t, 5, counter)
			close(done)
		}).Unhook,
	)

	eventTicker.StartTicker(id)

	has := eventTicker.HasTicker(id)
	require.True(t, has, "EventTicker.HasTicker() returned false, expected true")

	select {
	case <-done:
		unhook()
		eventTicker.Shutdown()
	case <-time.After(10 * time.Second):
		t.Error("Shutdown timed out")
	}
}

func TestRescheduleMultipleTicker(t *testing.T) {
	eventTicker := eventticker.New(
		eventticker.RetryInterval[index, testID](time.Duration(2*time.Second)),
		eventticker.RetryJitter[index, testID](time.Duration(time.Millisecond)),
		eventticker.MaxRequestThreshold[index, testID](2),
	)

	type info struct {
		startTime time.Time
		counter   int
	}

	tickers := make(map[testID]*info)
	for i := 1; i <= 5; i++ {
		id := newtestID(index(i), randByte())
		tickers[id] = &info{}
	}

	var mutex sync.Mutex
	var wg sync.WaitGroup

	wg.Add(len(tickers))
	unhook := lo.Batch(
		eventTicker.Events.TickerStarted.Hook(func(ti testID) {
			mutex.Lock()
			defer mutex.Unlock()

			t.Logf("ticker %v started", ti)
			tickers[ti].startTime = time.Now()
		}).Unhook,
		eventTicker.Events.Tick.Hook(func(ti testID) {
			mutex.Lock()
			defer mutex.Unlock()

			t.Logf("ticker %v ticks", ti)
			if tickers[ti].counter > 0 {
				require.LessOrEqual(t, time.Since(tickers[ti].startTime), time.Duration(3*time.Second))
				tickers[ti].startTime = time.Now()
			}
			tickers[ti].counter++
		}).Unhook,
		eventTicker.Events.TickerFailed.Hook(func(ti testID) {
			t.Logf("ticker %v terminated", ti)
			require.Equal(t, 4, tickers[ti].counter)
			wg.Done()
		}).Unhook,
	)

	issuerIDs := lo.Keys(tickers)
	require.Equal(t, len(tickers), len(issuerIDs))
	eventTicker.StartTickers(issuerIDs)

	wg.Wait()
	unhook()
	eventTicker.Shutdown()
}

func TestNonExistedTicker(t *testing.T) {
	eventTicker := eventticker.New(
		eventticker.RetryInterval[index, testID](time.Duration(time.Second)),
		eventticker.RetryJitter[index, testID](time.Duration(time.Millisecond)),
	)

	id2 := newtestID(index(2), randByte())
	eventTicker.StartTicker(id2)

	// Test HasTicker with non-existed ticker ID
	id3 := newtestID(index(2), 3)
	has := eventTicker.HasTicker(id3)
	require.False(t, has, "EventTicker.HasTicker() returned true, expected false")
	eventTicker.StopTicker(id3)

	// Test StopTicker with non-existed slot index
	id4 := newtestID(index(3), 4)
	has = eventTicker.HasTicker(id4)
	require.False(t, has, "EventTicker.HasTicker() returned true, expected false")
	eventTicker.StopTicker(id4)
}

type testID [65]byte

func newtestID(i index, idByte byte) testID {
	id := testID{}
	copy(id[:], []byte{idByte})
	binary.LittleEndian.PutUint64(id[1:], uint64(i))

	return id
}

func (t testID) Bytes() ([]byte, error) {
	return t[:], nil
}

func (t testID) Index() index {
	return index(binary.LittleEndian.Uint64(t[1:]))
}

func (t testID) String() string {
	return base58.Encode(t[:])
}

type index uint32

func randByte() byte {
	return byte(rand.Intn(256))
}
