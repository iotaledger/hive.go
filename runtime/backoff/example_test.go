package backoff

import (
	"fmt"
	"time"
)

func ExampleRetry() {
	// An operation that may fail.
	operation := func() error {
		fmt.Println("do something")

		return nil // or an error
	}

	err := Retry(ExponentialBackOff(100*time.Millisecond, 1.5), operation)
	if err != nil {
		// Handle error.
		return
	}

	// Output: do something
}

func ExampleMaxRetries() {
	// An operation that may fail.
	operation := func() error {
		fmt.Println("do something")

		return errTest
	}

	p := ConstantBackOff(100 * time.Millisecond).With(MaxRetries(2))
	err := Retry(p, operation)
	if err != nil {
		// Handle error.
		return
	}

	// Output:
	// do something
	// do something
	// do something
}
