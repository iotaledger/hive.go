package eventticker_test

import (
	"encoding/binary"
	"testing"
	"time"

	"github.com/iotaledger/hive.go/core/eventticker"
	"github.com/mr-tron/base58"
	"github.com/stretchr/testify/require"
)

func TestEventTicker(t *testing.T) {
	eventTicker := eventticker.New[index, testID]()

	id := newtestID(index(1), 0)
	eventTicker.StartTicker(id)

	has := eventTicker.HasTicker(id)
	require.True(t, has, "EventTicker.HasTicker() returned false, expected true")
	require.Equalf(t, 1, eventTicker.QueueSize(), "EventTicker.QueueSize() returned %d, expected 1", eventTicker.QueueSize())

	eventTicker.StopTicker(id)
	has = eventTicker.HasTicker(id)
	require.False(t, has, "EventTicker.HasTicker() returned true, expected false")
	require.Equalf(t, 0, eventTicker.QueueSize(), "EventTicker.QueueSize() returned %d, expected 0", eventTicker.QueueSize())

	id1 := newtestID(index(1), 1)
	id2 := newtestID(index(2), 2)
	eventTicker.StartTickers([]testID{id1, id2})

	has = eventTicker.HasTicker(id1)
	require.True(t, has, "EventTicker.HasTicker() returned false, expected true")
	has = eventTicker.HasTicker(id2)
	require.True(t, has, "EventTicker.HasTicker() returned false, expected true")

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
		t.Errorf("Shutdown timed out")
	}
}

type testID [65]byte

func newtestID(i index, idByte byte) testID {
	id := testID{}
	copy(id[:], []byte{idByte})
	binary.LittleEndian.PutUint64(id[0:], uint64(i))

	return id
}

func (t testID) Bytes() ([]byte, error) {
	return t[:], nil
}

func (t *testID) FromBytes(b []byte) (int, error) {
	copy(t[:], b)
	return len(t), nil
}

func (t testID) Index() index {
	return index(binary.LittleEndian.Uint64(t[:63]))
}

func (t testID) String() string {
	return base58.Encode(t[:])
}

type index uint64
