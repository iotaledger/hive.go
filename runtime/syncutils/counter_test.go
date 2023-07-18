package syncutils_test

import (
	"sync"
	"testing"

	"github.com/iotaledger/hive.go/runtime/syncutils"
)

func TestCounter_IncreaseDecrease(t *testing.T) {
	counter := syncutils.NewCounter()
	var wg sync.WaitGroup

	// Test parallel increase and decrease
	for i := 0; i < 1000000; i++ {
		wg.Add(2)
		go func() {
			defer wg.Done()
			counter.Increase()
		}()
		go func() {
			defer wg.Done()
			counter.Decrease()
		}()
	}

	wg.Wait()

	if val := counter.Get(); val != 0 {
		t.Errorf("Expected: 0, Got: %d", val)
	}
}

func TestCounter_WaitIsAboveBelow(t *testing.T) {
	counter := syncutils.NewCounter()
	var wg sync.WaitGroup

	// Test parallel waits
	wg.Add(1)
	go func() {
		defer wg.Done()
		counter.WaitIsAbove(500)
		if val := counter.Get(); val <= 500 {
			t.Error("Value is not above 500")
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 1000; i++ {
			counter.Increase()
		}
	}()

	wg.Wait()
}

func TestCounter_Subscribe(t *testing.T) {
	counter := syncutils.NewCounter()
	var wg sync.WaitGroup

	subscription := func(oldValue, newValue int) {
		t.Logf("Value changed from %d to %d", oldValue, newValue)
	}

	unsubscribe := counter.Subscribe(subscription)

	// Test parallel update and subscription notification
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 1000; i++ {
			counter.Increase()
		}
	}()

	wg.Wait()

	// Test unsubscribe
	unsubscribe()
}

func TestCounter_WaitIsBelowZero(t *testing.T) {
	counter := syncutils.NewCounter()

	// Set initial value
	counter.Set(1000)

	var wg sync.WaitGroup

	// Spawn goroutines that will wait until counter is below a certain threshold
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			counter.WaitIsBelow(200)
			if val := counter.Get(); val >= 200 {
				t.Errorf("Expected value below 200, got: %d", val)
			}
		}()
	}

	// Spawn goroutines that will wait until counter is zero
	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			counter.WaitIsZero()
			if val := counter.Get(); val != 0 {
				t.Errorf("Expected value to be zero, got: %d", val)
			}
		}()
	}

	// Spawn goroutines that decrease the counter value
	wg.Add(1)
	go func() {
		defer wg.Done()
		for counter.Decrease() > 0 {
		}
	}()

	wg.Wait()
}
