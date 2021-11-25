package filter

import (
	"github.com/iotaledger/hive.go/v2/syncutils"
)

type ByteArrayFilter struct {
	byteArrays      [][]byte
	byteArraysByKey map[string]bool
	size            int
	mutex           syncutils.RWMutex
}

func NewByteArrayFilter(size int) *ByteArrayFilter {
	return &ByteArrayFilter{
		byteArrays:      make([][]byte, 0, size),
		byteArraysByKey: make(map[string]bool, size),
		size:            size,
	}
}

func (filter *ByteArrayFilter) Contains(byteArray []byte) bool {
	filter.mutex.RLock()
	defer filter.mutex.RUnlock()

	_, exists := filter.byteArraysByKey[string(byteArray)]

	return exists
}

func (filter *ByteArrayFilter) Add(byteArray []byte) bool {
	key := string(byteArray)

	filter.mutex.Lock()
	defer filter.mutex.Unlock()

	if _, exists := filter.byteArraysByKey[key]; !exists {
		if len(filter.byteArrays) == filter.size {
			delete(filter.byteArraysByKey, string(filter.byteArrays[0]))

			filter.byteArrays = append(filter.byteArrays[1:], byteArray)
		} else {
			filter.byteArrays = append(filter.byteArrays, byteArray)
		}

		filter.byteArraysByKey[key] = true

		return true
	} else {
		return false
	}
}
