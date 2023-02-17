package workerpool

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func Test(t *testing.T) {
	group := NewGroup(t.Name())
	_ = group.CreatePool("poolA")

	require.Equal(t, group, group.Root())

	subgroup1 := group.CreateGroup("sub1")
	pool1 := subgroup1.CreatePool("pool1")
	pool2 := subgroup1.CreatePool("pool2")

	subgroup2 := group.CreateGroup("sub2")
	subSubGroup := subgroup2.CreateGroup("loop")
	_ = subSubGroup.CreatePool("pool3")

	require.Equal(t, group, subSubGroup.Root())

	pool1.Submit(func() {
		time.Sleep(1 * time.Second)

		fmt.Println("TASK1 done")
	})

	pool2.Submit(func() {
		time.Sleep(3 * time.Second)

		fmt.Println("TASK2 done")
	})

	fmt.Println(group)
	fmt.Println(group.Pools())

	group.Shutdown()

	fmt.Println("ALL TASKS DONE")
}

func TestRecreateGroup(t *testing.T) {
	workers := NewGroup(t.Name())

	workers.CreateGroup("group1").Shutdown()
	workers.CreateGroup("group1")

	require.Panics(t, func() {
		workers.CreateGroup("group1")
		workers.CreateGroup("group1")
	})
}

func TestRecreatePool(t *testing.T) {
	workers := NewGroup(t.Name())

	workers.CreatePool("pool1").Shutdown()
	workers.CreatePool("pool1")

	require.Panics(t, func() {
		workers.CreatePool("pool1")
		workers.CreatePool("pool1")
	})
}
