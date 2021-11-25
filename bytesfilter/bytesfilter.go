package bytesfilter

import (
	"sync"

	"github.com/iotaledger/hive.go/types"
)

type BytesFilter struct {
	byteArrays [][]byte
	bytesByKey map[string]types.Empty
	size       int
	mutex      sync.RWMutex
}

func New(size int) *BytesFilter {
	return &BytesFilter{
		byteArrays: make([][]byte, 0, size),
		bytesByKey: make(map[string]types.Empty, size),
		size:       size,
	}
}

func (bytesFilter *BytesFilter) Add(bytes []byte) bool {
	key := string(bytes)

	bytesFilter.mutex.Lock()

	if _, exists := bytesFilter.bytesByKey[key]; !exists {
		if len(bytesFilter.byteArrays) == bytesFilter.size {
			delete(bytesFilter.bytesByKey, string(bytesFilter.byteArrays[0]))

			bytesFilter.byteArrays[0] = nil
			bytesFilter.byteArrays = append(bytesFilter.byteArrays[1:], bytes)
		} else {
			bytesFilter.byteArrays = append(bytesFilter.byteArrays, bytes)
		}

		bytesFilter.bytesByKey[key] = types.Void

		bytesFilter.mutex.Unlock()

		return true
	} else {
		bytesFilter.mutex.Unlock()

		return false
	}
}

func (bytesFilter *BytesFilter) Contains(byteArray []byte) (exists bool) {
	bytesFilter.mutex.RLock()
	_, exists = bytesFilter.bytesByKey[string(byteArray)]
	bytesFilter.mutex.RUnlock()

	return
}
