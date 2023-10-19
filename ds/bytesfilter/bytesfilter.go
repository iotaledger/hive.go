package bytesfilter

import (
	"sync"

	"github.com/iotaledger/hive.go/ds/shrinkingmap"
	"github.com/iotaledger/hive.go/ds/types"
)

type BytesFilter[IdentifierType types.IdentifierType] struct {
	knownIdentifiers *shrinkingmap.ShrinkingMap[IdentifierType, types.Empty]
	identifiers      []IdentifierType

	newIdentifierFunc func([]byte) IdentifierType
	size              int
	mutex             sync.RWMutex
}

func New[IdentifierType types.IdentifierType](newIdentifierFunc func([]byte) IdentifierType, size int) *BytesFilter[IdentifierType] {
	return &BytesFilter[IdentifierType]{
		knownIdentifiers:  shrinkingmap.New[IdentifierType, types.Empty](shrinkingmap.WithShrinkingThresholdCount(size)),
		identifiers:       make([]IdentifierType, 0, size),
		newIdentifierFunc: newIdentifierFunc,
		size:              size,
	}
}

func (b *BytesFilter[IdentifierType]) Add(bytes []byte) (identifier IdentifierType, added bool) {
	identifier = b.newIdentifierFunc(bytes)

	b.mutex.Lock()
	defer b.mutex.Unlock()

	return identifier, b.addIdentifier(identifier)
}

func (b *BytesFilter[IdentifierType]) AddIdentifier(identifier IdentifierType) (added bool) {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	return b.addIdentifier(identifier)
}

func (b *BytesFilter[IdentifierType]) Contains(bytes []byte) (exists bool) {
	b.mutex.RLock()
	defer b.mutex.RUnlock()

	_, exists = b.knownIdentifiers.Get(b.newIdentifierFunc(bytes))

	return exists
}

func (b *BytesFilter[IdentifierType]) ContainsIdentifier(identifier IdentifierType) (exists bool) {
	b.mutex.RLock()
	defer b.mutex.RUnlock()

	_, exists = b.knownIdentifiers.Get(identifier)

	return exists
}

func (b *BytesFilter[IdentifierType]) addIdentifier(identifier IdentifierType) (added bool) {
	if _, exists := b.knownIdentifiers.Get(identifier); exists {
		return false
	}

	if len(b.identifiers) == b.size {
		b.knownIdentifiers.Delete(b.identifiers[0])
		b.identifiers = append(b.identifiers[1:], identifier)
	} else {
		b.identifiers = append(b.identifiers, identifier)
	}

	b.knownIdentifiers.Set(identifier, types.Void)

	return true
}
