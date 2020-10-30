package redblacktree

import (
	"fmt"
	"math/rand"
	"strconv"
	"testing"
	"time"
)

func TestTree_Put(t *testing.T) {
	tree := New()

	numbersToAdd := make([]string, 30)
	for i := 0; i < len(numbersToAdd); i++ {
		numbersToAdd[i] = "key" + strconv.Itoa(i)
	}
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(numbersToAdd), func(i, j int) { numbersToAdd[i], numbersToAdd[j] = numbersToAdd[j], numbersToAdd[i] })

	for _, i := range numbersToAdd {
		tree.Set(i, i)
	}

	fmt.Println(tree.Keys())

	for i := 5; i < 15; i++ {
		tree.Delete("key" + strconv.Itoa(i))
	}

	fmt.Println(tree.Keys())

	fmt.Println(tree.Get("key11"))
	fmt.Println(tree.Get("key2"))
	fmt.Println(tree.Get("key3"))
}
