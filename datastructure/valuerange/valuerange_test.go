package valuerange

import (
	"fmt"
	"testing"
)

func TestValueRange_Compare(t *testing.T) {
	valueRange0 := All()
	fmt.Println(valueRange0)

	valueRangeAtMost := AtMost(Int64Value(100))
	fmt.Println(valueRangeAtMost)

	valueRange1 := Open(Int64Value(10), Int64Value(14))
	fmt.Println(valueRange1)

	valueRange2 := Closed(Int64Value(10), Int64Value(14))
	fmt.Println(valueRange2)

	valueRange3 := GreaterThan(Int64Value(10))
	fmt.Println(valueRange3)

	fmt.Print(valueRange1.Contains(Int64Value(13)))
}
