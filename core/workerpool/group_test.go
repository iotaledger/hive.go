package workerpool

import (
	"fmt"
	"testing"
	"time"
)

func Test(t *testing.T) {
	group := NewGroup("protocol")
	_ = group.CreatePool("poolA")

	subgroup1 := group.CreateGroup("sub1")
	pool1 := subgroup1.CreatePool("pool1")
	pool2 := subgroup1.CreatePool("pool2")

	subgroup2 := group.CreateGroup("sub2")
	subSubGroup := subgroup2.CreateGroup("loop")
	_ = subSubGroup.CreatePool("pool3")

	pool1.Submit(func() {
		time.Sleep(1 * time.Second)

		fmt.Println("TASK1 done")
	})

	pool2.Submit(func() {
		time.Sleep(3 * time.Second)

		fmt.Println("TASK2 done")
	})

	fmt.Println(group)

	group.Shutdown()

	fmt.Println("ALL TASKS DONE")
}
