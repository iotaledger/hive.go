package test

import (
	"fmt"
	"testing"

	"github.com/iotaledger/hive.go/objectstorage"
)

func TestRetainedPartition(t *testing.T) {
	retainedPartition := objectstorage.NewRetainedPartition()

	retainedPartition.Retain([]string{"A", "B"})
	retainedPartition.Release([]string{"A", "B"})

	fmt.Println(retainedPartition.IsRetained([]string{"A", "B"}))
}
