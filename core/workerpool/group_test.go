package workerpool

import (
	"fmt"
	"testing"
	"time"
)

func Test(t *testing.T) {
	group := NewGroup("protocol")

	pool1 := group.CreatePool("pool1")
	pool2 := group.CreatePool("pool2")

	pool1.Submit(func() {
		time.Sleep(1 * time.Second)

		fmt.Println("TASK1 done")
	})

	pool2.Submit(func() {
		time.Sleep(3 * time.Second)

		fmt.Println("TASK2 done")
	})

	group.Shutdown()

	fmt.Println("ALL TASKS DONE")
}
