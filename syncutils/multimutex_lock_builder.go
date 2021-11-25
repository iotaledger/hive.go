package syncutils

import (
	"github.com/iotaledger/hive.go/datastructure/set"
)

type MultiMutexLockBuilder struct {
	locks           []interface{}
	seenIdentifiers set.Set
}

func (lockBuilder *MultiMutexLockBuilder) AddLock(identifier interface{}) *MultiMutexLockBuilder {
	if lockBuilder.seenIdentifiers == nil {
		lockBuilder.seenIdentifiers = set.New()
	}

	if lockBuilder.seenIdentifiers.Add(identifier) {
		lockBuilder.locks = append(lockBuilder.locks, identifier)
	}

	return lockBuilder
}

func (lockBuilder *MultiMutexLockBuilder) Build() []interface{} {
	return lockBuilder.locks
}
