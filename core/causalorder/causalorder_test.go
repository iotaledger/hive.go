package causalorder

import (
	"math/rand"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/hive.go/runtime/workerpool"
)

// TestCausalOrder_Queue tests the queueing of entities in the CausalOrder.
func TestCausalOrder_Queue(t *testing.T) {
	workers := workerpool.NewGroup(t.Name())
	tf := NewTestFramework(t, workers)

	tf.CreateEntity("A", WithParents(tf.EntityIDs("Genesis")), WithIndex(1))
	tf.CreateEntity("B", WithParents(tf.EntityIDs("A")), WithIndex(1))
	tf.CreateEntity("C", WithParents(tf.EntityIDs("A", "B")), WithIndex(1))
	tf.CreateEntity("D", WithParents(tf.EntityIDs("C", "B")), WithIndex(1))
	tf.CreateEntity("E", WithParents(tf.EntityIDs("C", "D")), WithIndex(1))

	tf.Queue(tf.Entity("A"))
	workers.WaitChildren()
	tf.AssertOrdered("A")

	tf.Queue(tf.Entity("A"))
	workers.WaitChildren()
	tf.AssertOrdered("A")

	tf.Queue(tf.Entity("D"))
	workers.WaitChildren()
	tf.AssertOrdered("A")

	tf.Queue(tf.Entity("E"))
	workers.WaitChildren()
	tf.AssertOrdered("A")

	tf.Queue(tf.Entity("C"))
	workers.WaitChildren()
	tf.AssertOrdered("A")

	tf.Queue(tf.Entity("B"))
	workers.WaitChildren()
	tf.AssertOrdered("A", "B", "C", "D", "E")
}

// TestCausalOrder_EvictSlot tests the eviction of entities in the CausalOrder.
func TestCausalOrder_EvictSlot(t *testing.T) {
	workers := workerpool.NewGroup(t.Name())
	tf := NewTestFramework(t, workers)
	tf.CreateEntity("A", WithParents(tf.EntityIDs("Genesis")), WithIndex(1))
	tf.CreateEntity("B", WithParents(tf.EntityIDs("A")), WithIndex(1))
	tf.CreateEntity("C", WithParents(tf.EntityIDs("A", "B")), WithIndex(1))
	tf.CreateEntity("D", WithParents(tf.EntityIDs("C", "B")), WithIndex(1))
	tf.CreateEntity("E", WithParents(tf.EntityIDs("C", "D")), WithIndex(1))
	tf.CreateEntity("F", WithParents(tf.EntityIDs("Genesis")), WithIndex(1))
	tf.CreateEntity("G", WithParents(tf.EntityIDs("F")), WithIndex(1))
	tf.CreateEntity("H", WithParents(tf.EntityIDs("G")), WithIndex(2))

	tf.Queue(tf.Entity("A"))
	workers.WaitChildren()
	tf.AssertOrdered("A")
	tf.AssertEvicted()

	tf.Queue(tf.Entity("D"))
	workers.WaitChildren()
	tf.AssertOrdered("A")
	tf.AssertEvicted()

	tf.Queue(tf.Entity("E"))
	workers.WaitChildren()
	tf.AssertOrdered("A")
	tf.AssertEvicted()

	tf.Queue(tf.Entity("C"))
	workers.WaitChildren()
	tf.AssertOrdered("A")
	tf.AssertEvicted()

	tf.Queue(tf.Entity("B"))
	workers.WaitChildren()
	tf.AssertOrdered("A", "B", "C", "D", "E")
	tf.AssertEvicted()

	tf.Queue(tf.Entity("G"))
	workers.WaitChildren()
	tf.AssertOrdered("A", "B", "C", "D", "E")
	tf.AssertEvicted()

	tf.EvictIndex(1)
	workers.WaitChildren()
	tf.AssertOrdered("A", "B", "C", "D", "E")
	tf.AssertEvicted("F", "G")

	tf.Queue(tf.Entity("F"))
	workers.WaitChildren()
	tf.AssertOrdered("A", "B", "C", "D", "E")
	tf.AssertEvicted("F", "G")

	tf.Queue(tf.Entity("H"))
	workers.WaitChildren()
	tf.AssertOrdered("A", "B", "C", "D", "E")
	tf.AssertEvicted("F", "G", "H")
}

// TestCausalOrder_UnexpectedCases tests the unexpected cases of the CausalOrder.
func TestCausalOrder_UnexpectedCases(t *testing.T) {
	workers := workerpool.NewGroup(t.Name())
	tf := NewTestFramework(t, workers)
	tf.CreateEntity("A", WithParents(tf.EntityIDs("Genesis")), WithIndex(1))
	tf.CreateEntity("B", WithParents(tf.EntityIDs("A")), WithIndex(1))
	tf.CreateEntity("C", WithParents(tf.EntityIDs("A")), WithIndex(1))
	tf.Queue(tf.Entity("C"))

	// test queueing an entity with non-existing parents
	tf.RemoveEntity("A")
	tf.Queue(tf.Entity("B"))
	workers.WaitChildren()
	tf.AssertOrdered()
	tf.AssertEvicted("B")

	// test eviction of non-existing entity
	tf.RemoveEntity("C")
	defer func() {
		require.NotNil(t, recover())
		workers.WaitChildren()
		tf.AssertOrdered()
		tf.AssertEvicted("B")
	}()
	tf.EvictIndex(1)
}

func TestCausalOrder_QueueParallel(t *testing.T) {
	workers := workerpool.NewGroup(t.Name())
	tf := NewTestFramework(t, workers)
	var wg sync.WaitGroup

	aliases := map[string]bool{
		"A": true,
		"B": true,
		"C": false,
		"D": true,
		"E": true,
	}

	tf.CreateEntity("A", WithParents(tf.EntityIDs("Genesis")), WithIndex(1))
	tf.CreateEntity("B", WithParents(tf.EntityIDs("A")), WithIndex(1))
	tf.CreateEntity("C", WithParents(tf.EntityIDs("A", "B")), WithIndex(1))
	tf.CreateEntity("D", WithParents(tf.EntityIDs("C", "B")), WithIndex(1))
	tf.CreateEntity("E", WithParents(tf.EntityIDs("D")), WithIndex(2))

	for alias, queue := range aliases {
		wg.Add(1)
		go func(alias string, queue bool) {
			if queue {
				tf.Queue(tf.Entity(alias))
			}
			wg.Done()
		}(alias, queue)
	}

	wg.Wait()
	workers.WaitChildren()
	tf.EvictIndex(1)
	tf.AssertOrdered("A", "B")
	tf.AssertEvicted("C", "D")

	tf.EvictUntil(2)
	tf.AssertOrdered("A", "B")
	tf.AssertEvicted("C", "D", "E")
}

func TestCausalOrder_QueueParallelMassive(t *testing.T) {
	var rand = rand.New(rand.NewSource(time.Now().UnixNano()))
	var wg sync.WaitGroup

	workers := workerpool.NewGroup(t.Name())
	tf := NewTestFramework(t, workers)

	// generate bunch of IDs
	count := 10000
	aliases := make([]string, count)
	seen := make(map[string]bool)

	for i := 0; i < count; i++ {
		id := randSeq(10)
		if _, exist := seen[id]; exist {
			i--
			continue
		}
		seen[id] = true
		aliases[i] = id
	}
	require.Equal(t, count, len(aliases))

	tf.CreateEntity(aliases[0], WithParents(tf.EntityIDs("Genesis")), WithIndex(1))
	for i := 1; i < count; i++ {
		numParents := i % 20
		parents := make([]string, 0)
		for j := 0; j < numParents; j++ {
			p := rand.Intn(i)
			parents = append(parents, aliases[p])
		}
		tf.CreateEntity(aliases[i], WithParents(tf.EntityIDs(parents...)), WithIndex(1))
	}

	wg.Add(count)
	for _, alias := range aliases {
		go func(alias string) {
			tf.Queue(tf.Entity(alias))
			wg.Done()
		}(alias)
	}

	wg.Wait()
	workers.WaitChildren()

	tf.AssertOrdered(aliases...)
}

func TestCausalOrder_EvictParallel(t *testing.T) {
	workers := workerpool.NewGroup(t.Name())
	tf := NewTestFramework(t, workers)
	var wg sync.WaitGroup

	tf.CreateEntity("A", WithParents(tf.EntityIDs("Genesis")), WithIndex(1))
	tf.CreateEntity("B", WithParents(tf.EntityIDs("A")), WithIndex(1))
	tf.CreateEntity("C", WithParents(tf.EntityIDs("A", "B")), WithIndex(1))
	tf.CreateEntity("D", WithParents(tf.EntityIDs("C", "B")), WithIndex(1))
	tf.CreateEntity("E", WithParents(tf.EntityIDs("D")), WithIndex(2))

	wg.Wait()
	tf.Queue(tf.Entity("A"))
	workers.WaitChildren()
	tf.AssertOrdered("A")

	tf.Queue(tf.Entity("D"))
	workers.WaitChildren()
	tf.AssertOrdered("A")

	tf.Queue(tf.Entity("E"))
	workers.WaitChildren()
	tf.AssertOrdered("A")

	tf.Queue(tf.Entity("B"))
	workers.WaitChildren()
	tf.AssertOrdered("A", "B")
	tf.AssertEvicted()

	// tf.EvictUntil(2)
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(tf *TestFramework, index uint32) {
			tf.EvictUntil(index)
			wg.Done()
		}(tf, uint32(i))
	}
	wg.Wait()

	tf.AssertEvicted("C", "D", "E")
}

func randSeq(n int) string {
	letters := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}
