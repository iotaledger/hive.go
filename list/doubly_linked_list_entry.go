package list

import (
	"github.com/iotaledger/hive.go/v2/syncutils"
)

type DoublyLinkedListEntry struct {
	Value interface{}
	Prev  *DoublyLinkedListEntry
	Next  *DoublyLinkedListEntry
	mutex syncutils.RWMutex
}

func (entry *DoublyLinkedListEntry) GetNext() *DoublyLinkedListEntry {
	entry.mutex.RLock()
	defer entry.mutex.RUnlock()

	return entry.Next
}

func (entry *DoublyLinkedListEntry) SetNext(next *DoublyLinkedListEntry) {
	entry.mutex.Lock()
	defer entry.mutex.Unlock()

	entry.Next = next
}

func (entry *DoublyLinkedListEntry) GetPrev() *DoublyLinkedListEntry {
	entry.mutex.RLock()
	defer entry.mutex.RUnlock()

	return entry.Prev
}

func (entry *DoublyLinkedListEntry) SetPrev(prev *DoublyLinkedListEntry) {
	entry.mutex.Lock()
	defer entry.mutex.Unlock()

	entry.Prev = prev
}

func (entry *DoublyLinkedListEntry) GetValue() interface{} {
	entry.mutex.RLock()
	defer entry.mutex.RUnlock()

	return entry.Value
}

func (entry *DoublyLinkedListEntry) SetValue(value interface{}) {
	entry.mutex.Lock()
	defer entry.mutex.Unlock()

	entry.Value = value
}
