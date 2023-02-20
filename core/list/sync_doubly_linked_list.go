package list

import (
	"github.com/iotaledger/hive.go/runtime/syncutils"
)

func NewSyncDoublyLinkedList() *SyncDoublyLinkedList {
	return &SyncDoublyLinkedList{}
}

// SyncDoublyLinkedList is a DoublyLinkedList but with synchronized methods.
type SyncDoublyLinkedList struct {
	mutex  syncutils.RWMutex
	Unsafe DoublyLinkedList
}

// Appends the specified value to the end of this list.
func (list *SyncDoublyLinkedList) Add(value interface{}) *DoublyLinkedListEntry {
	return list.AddLast(value)
}

// Appends the specified element to the end of this list.
func (list *SyncDoublyLinkedList) AddEntry(entry *DoublyLinkedListEntry) {
	list.AddLastEntry(entry)
}

func (list *SyncDoublyLinkedList) AddLast(value interface{}) *DoublyLinkedListEntry {
	newEntry := &DoublyLinkedListEntry{Value: value}
	list.AddLastEntry(newEntry)

	return newEntry
}

func (list *SyncDoublyLinkedList) AddFirst(value interface{}) *DoublyLinkedListEntry {
	newEntry := &DoublyLinkedListEntry{Value: value}
	list.AddFirstEntry(newEntry)

	return newEntry
}

func (list *SyncDoublyLinkedList) GetFirst() (interface{}, error) {
	firstEntry, err := list.GetFirstEntry()
	if err != nil {
		return nil, err
	}

	return firstEntry.GetValue(), nil
}

func (list *SyncDoublyLinkedList) GetLast() (interface{}, error) {
	lastEntry, err := list.GetLastEntry()
	if err != nil {
		return nil, err
	}

	return lastEntry.GetValue(), nil
}

func (list *SyncDoublyLinkedList) RemoveFirst() (interface{}, error) {
	firstEntry, err := list.RemoveFirstEntry()
	if err != nil {
		return nil, err
	}

	return firstEntry.GetValue(), nil
}

func (list *SyncDoublyLinkedList) RemoveLast() (interface{}, error) {
	lastEntry, err := list.RemoveLastEntry()
	if err != nil {
		return nil, err
	}

	return lastEntry.GetValue(), nil
}

func (list *SyncDoublyLinkedList) AddLastEntry(entry *DoublyLinkedListEntry) {
	list.mutex.Lock()
	defer list.mutex.Unlock()
	list.Unsafe.AddLastEntry(entry)
}

func (list *SyncDoublyLinkedList) AddFirstEntry(entry *DoublyLinkedListEntry) {
	list.mutex.Lock()
	defer list.mutex.Unlock()
	list.Unsafe.AddFirstEntry(entry)
}

func (list *SyncDoublyLinkedList) GetFirstEntry() (*DoublyLinkedListEntry, error) {
	list.mutex.RLock()
	defer list.mutex.RUnlock()

	return list.Unsafe.GetFirstEntry()
}

func (list *SyncDoublyLinkedList) GetLastEntry() (*DoublyLinkedListEntry, error) {
	list.mutex.RLock()
	defer list.mutex.RUnlock()

	return list.Unsafe.GetLastEntry()
}

func (list *SyncDoublyLinkedList) RemoveFirstEntry() (*DoublyLinkedListEntry, error) {
	list.mutex.Lock()
	defer list.mutex.Unlock()

	return list.Unsafe.RemoveFirstEntry()
}

func (list *SyncDoublyLinkedList) RemoveLastEntry() (*DoublyLinkedListEntry, error) {
	list.mutex.Lock()
	defer list.mutex.Unlock()

	return list.Unsafe.RemoveLastEntry()
}

func (list *SyncDoublyLinkedList) RemoveEntry(entry *DoublyLinkedListEntry) error {
	list.mutex.Lock()
	defer list.mutex.Unlock()

	return list.Unsafe.RemoveEntry(entry)
}

func (list *SyncDoublyLinkedList) Remove(value interface{}) error {
	list.mutex.RLock()
	currentEntry := list.Unsafe.head
	for currentEntry != nil {
		if currentEntry.GetValue() == value {
			list.mutex.RUnlock()

			if err := list.RemoveEntry(currentEntry); err != nil {
				return err
			}

			return nil
		}

		currentEntry = currentEntry.GetNext()
	}
	list.mutex.RUnlock()

	return ErrNoSuchElement
}

func (list *SyncDoublyLinkedList) Clear() {
	list.mutex.Lock()
	defer list.mutex.Unlock()
	list.Unsafe.Clear()
}

func (list *SyncDoublyLinkedList) GetSize() int {
	list.mutex.RLock()
	defer list.mutex.RUnlock()

	return list.Unsafe.count
}
