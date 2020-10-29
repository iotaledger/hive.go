package redblacktree

import (
	"fmt"
	"math/rand"
	"testing"
	"time"
)

func TestTree_Put(t *testing.T) {
	tree := New(IntComparator)

	numbersToAdd := make([]int, 30)
	for i := 0; i < len(numbersToAdd); i++ {
		numbersToAdd[i] = i
	}
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(numbersToAdd), func(i, j int) { numbersToAdd[i], numbersToAdd[j] = numbersToAdd[j], numbersToAdd[i] })

	for _, i := range numbersToAdd {
		tree.Put(i, i)
	}

	fmt.Println(tree)

	fmt.Println(tree.Get(1))
	fmt.Println(tree.Get(2))
	fmt.Println(tree.Get(3))
}

func IntComparator(a, b interface{}) int {
	aAsserted := a.(int)
	bAsserted := b.(int)
	switch {
	case aAsserted > bAsserted:
		return 1
	case aAsserted < bAsserted:
		return -1
	default:
		return 0
	}
}
