package workerpool

import (
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_SimpleCounter(t *testing.T) {
	const queueSize = 10
	const incCount = 100

	el := NewEventLoop(QueueSize(queueSize))

	var counter uint64
	incAtomic := func() {
		atomic.AddUint64(&counter, 1)
	}

	for i := 0; i < incCount; i++ {
		added := el.TrySubmit(incAtomic)

		if i < queueSize {
			assert.True(t, added)
		} else {
			assert.False(t, added)
		}
	}

	el.Start()

	for i := 0; i < incCount-queueSize; i++ {
		el.Submit(incAtomic)
	}

	el.StopAndWait()

	assert.Equal(t, uint64(incCount), counter)
}
