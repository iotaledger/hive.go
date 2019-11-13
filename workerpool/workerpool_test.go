package workerpool_test

import (
	"github.com/iotaledger/hive.go/workerpool"
	"sync"
	"testing"
)

func TestSignalBarrier(t *testing.T) {
	pool := workerpool.New(func(task workerpool.Task) {
		println(task.Param(0).(int))
		task.Return(nil)
	}, workerpool.WorkerCount(10), workerpool.QueueSize(2000))
	pool.Start()

	var wg sync.WaitGroup
	for i := 0; i < 200; i++ {
		wg.Add(1)

		go func(i int) {
			<-pool.Submit(i)

			wg.Done()
		}(i)
	}

	pool.SubmitBarrier()

	for i := 0; i < 200; i++ {
		wg.Add(1)

		go func(i int) {
			<-pool.Submit(i)

			wg.Done()
		}(i)
	}

	wg.Wait()

}

func Benchmark(b *testing.B) {
	pool := workerpool.New(func(task workerpool.Task) {
		task.Return(task.Param(0))
	}, workerpool.WorkerCount(10), workerpool.QueueSize(2000))
	pool.Start()

	var wg sync.WaitGroup
	for i := 0; i < b.N; i++ {
		wg.Add(1)

		go func(i int) {
			<-pool.Submit(i)

			wg.Done()
		}(i)
	}

	wg.Wait()
}
