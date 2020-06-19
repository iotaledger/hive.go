package list

import (
	"errors"
	"fmt"
)

var (
	ErrNoSuchElement   = errors.New("element does not exist")
	ErrInvalidArgument = errors.New("invalid argument")
)

func NewDoublyLinkedList() *DoublyLinkedList {
	return &DoublyLinkedList{}
}

type DoublyLinkedList struct {
	head  *DoublyLinkedListEntry
	tail  *DoublyLinkedListEntry
	count int
}

// Appends the specified value to the end of this list.
func (list *DoublyLinkedList) Add(value interface{}) *DoublyLinkedListEntry {
	return list.AddLast(value)
}

// Appends the specified element to the end of this list.
func (list *DoublyLinkedList) AddEntry(entry *DoublyLinkedListEntry) {
	list.AddLastEntry(entry)
}

func (list *DoublyLinkedList) AddLast(value interface{}) *DoublyLinkedListEntry {
	newEntry := &DoublyLinkedListEntry{Value: value}
	list.AddLastEntry(newEntry)
	return newEntry
}

func (list *DoublyLinkedList) AddFirst(value interface{}) *DoublyLinkedListEntry {
	newEntry := &DoublyLinkedListEntry{Value: value}
	list.AddFirstEntry(newEntry)
	return newEntry
}

func (list *DoublyLinkedList) GetFirst() (interface{}, error) {
	if firstEntry, err := list.GetFirstEntry(); err != nil {
		return nil, err
	} else {
		return firstEntry.GetValue(), nil
	}
}

func (list *DoublyLinkedList) GetLast() (interface{}, error) {
	if lastEntry, err := list.GetLastEntry(); err != nil {
		return nil, err
	} else {
		return lastEntry.GetValue(), nil
	}
}

func (list *DoublyLinkedList) RemoveFirst() (interface{}, error) {
	if firstEntry, err := list.RemoveFirstEntry(); err != nil {
		return nil, err
	} else {
		return firstEntry.GetValue(), nil
	}
}

func (list *DoublyLinkedList) RemoveLast() (interface{}, error) {
	if lastEntry, err := list.RemoveLastEntry(); err != nil {
		return nil, err
	} else {
		return lastEntry.GetValue(), nil
	}
}

func (list *DoublyLinkedList) AddLastEntry(entry *DoublyLinkedListEntry) {
	if list.head == nil {
		list.head = entry
	} else {
		list.tail.SetNext(entry)
		entry.SetPrev(list.tail)
	}

	list.tail = entry
	list.count++
}

func (list *DoublyLinkedList) AddFirstEntry(entry *DoublyLinkedListEntry) {
	if list.tail == nil {
		list.tail = entry
	} else {
		list.head.SetPrev(entry)
		entry.SetNext(list.head)
	}

	list.head = entry
	list.count++
}

func (list *DoublyLinkedList) GetFirstEntry() (*DoublyLinkedListEntry, error) {
	if list.head == nil {
		return nil, ErrNoSuchElement
	}
	return list.head, nil
}

func (list *DoublyLinkedList) GetLastEntry() (*DoublyLinkedListEntry, error) {
	if list.tail == nil {
		return nil, ErrNoSuchElement
	}
	return list.tail, nil
}

func (list *DoublyLinkedList) RemoveFirstEntry() (*DoublyLinkedListEntry, error) {
	entryToRemove := list.head
	if err := list.RemoveEntry(entryToRemove); err != nil {
		return nil, err
	}
	return entryToRemove, nil
}

func (list *DoublyLinkedList) RemoveLastEntry() (*DoublyLinkedListEntry, error) {
	entryToRemove := list.tail
	if err := list.RemoveEntry(entryToRemove); err != nil {
		return nil, err
	}
	return entryToRemove, nil
}

func (list *DoublyLinkedList) RemoveEntry(entry *DoublyLinkedListEntry) error {
	if entry == nil {
		return fmt.Errorf("%w: the entry must not be nil", ErrInvalidArgument)
	}

	if list.head == nil {
		return fmt.Errorf("%w: the entry is not part of the list", ErrNoSuchElement)
	}

	prevEntry := entry.GetPrev()
	nextEntry := entry.GetNext()

	if nextEntry != nil {
		nextEntry.SetPrev(prevEntry)
	}
	if list.head == entry {
		list.head = nextEntry
	}

	if prevEntry != nil {
		prevEntry.SetNext(nextEntry)
	}
	if list.tail == entry {
		list.tail = prevEntry
	}

	entry.SetNext(nil)
	entry.SetPrev(nil)

	list.count--

	return nil
}

func (list *DoublyLinkedList) Clear() {
	list.head = nil
	list.tail = nil
	list.count = 0
}

func (list *DoublyLinkedList) GetSize() int {
	return list.count
}
