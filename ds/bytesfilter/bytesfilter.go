package bytesfilter

import (
	"sync"

	"github.com/izuc/zipp.foundation/ds/shrinkingmap"
	"github.com/izuc/zipp.foundation/ds/types"
	dsTypes "github.com/izuc/zipp.foundation/ds/types"
)

type BytesFilter struct {
	knownIdentifiers *shrinkingmap.ShrinkingMap[types.Identifier, dsTypes.Empty]
	identifiers      []types.Identifier
	size             int
	mutex            sync.RWMutex
}

func New(size int) *BytesFilter {
	return &BytesFilter{
		knownIdentifiers: shrinkingmap.New[types.Identifier, dsTypes.Empty](shrinkingmap.WithShrinkingThresholdCount(size)),
		identifiers:      make([]types.Identifier, 0, size),
		size:             size,
	}
}

func (b *BytesFilter) Add(bytes []byte) (identifier types.Identifier, added bool) {
	identifier = types.NewIdentifier(bytes)

	b.mutex.Lock()
	defer b.mutex.Unlock()

	added = b.addIdentifier(identifier)

	return
}

func (b *BytesFilter) AddIdentifier(identifier types.Identifier) (added bool) {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	return b.addIdentifier(identifier)
}

func (b *BytesFilter) Contains(bytes []byte) (exists bool) {
	b.mutex.RLock()
	_, exists = b.knownIdentifiers.Get(types.NewIdentifier(bytes))
	b.mutex.RUnlock()

	return
}

func (b *BytesFilter) ContainsIdentifier(identifier types.Identifier) (exists bool) {
	b.mutex.RLock()
	_, exists = b.knownIdentifiers.Get(identifier)
	b.mutex.RUnlock()

	return
}

func (b *BytesFilter) addIdentifier(identifier types.Identifier) (added bool) {
	if _, exists := b.knownIdentifiers.Get(identifier); exists {
		return false
	}

	if len(b.identifiers) == b.size {
		b.knownIdentifiers.Delete(b.identifiers[0])
		b.identifiers = append(b.identifiers[1:], identifier)
	} else {
		b.identifiers = append(b.identifiers, identifier)
	}

	b.knownIdentifiers.Set(identifier, dsTypes.Void)

	return true
}
