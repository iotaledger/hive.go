package thresholdmap

import (
	"fmt"
	"testing"
)

func Test(t *testing.T) {
	thresholdMap := New(LowerThresholdMode)

	// marker two references marker 5
	thresholdMap.Set(2, 5)

	// marker 3 references marker 7
	thresholdMap.Set(3, 7)

	it := thresholdMap.Iterator()

	for it.HasPrev() {
		fmt.Println("ITERATORP", it.Prev().Value())
	}

	for it.HasNext() {
		fmt.Println("ITERATORN", it.Next().Value())
	}

	for it.HasPrev() {
		fmt.Println("ITERATORP", it.Prev().Value())
	}

	fmt.Println(thresholdMap.Get(1))
	fmt.Println(thresholdMap.Get(2))
	fmt.Println(thresholdMap.Get(3))
	fmt.Println(thresholdMap.Get(4))
	fmt.Println(thresholdMap.Get(99))
}
