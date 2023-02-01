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

	subGroup := group.CreateGroup("engine")
	pool3 := subGroup.CreatePool("booker")

	pool1.Submit(func() {
		time.Sleep(1 * time.Second)

		fmt.Println("TASK1 done")
	})

	pool2.Submit(func() {
		time.Sleep(3 * time.Second)

		fmt.Println("TASK2 done")
	})

	pool3.Submit(func() {
		time.Sleep(2 * time.Second)

		fmt.Println("TASK3 done")
	})

	go func() {
		for {
			fmt.Println(group)

			time.Sleep(500 * time.Millisecond)
		}
	}()

	group.Shutdown()

	fmt.Println("ALL TASKS DONE")

	time.Sleep(1 * time.Second)
}
