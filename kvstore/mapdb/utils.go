package mapdb

import (
	"sort"

	"github.com/iotaledger/hive.go/kvstore"
)

func sortSlice(slice []string, iterDirection ...kvstore.IterDirection) []string {

	switch kvstore.GetIterDirection(iterDirection...) {
	case kvstore.IterDirectionForward:
		sort.Sort(sort.StringSlice(slice))

	case kvstore.IterDirectionBackward:
		sort.Sort(sort.Reverse(sort.StringSlice(slice)))
	}

	return slice
}
