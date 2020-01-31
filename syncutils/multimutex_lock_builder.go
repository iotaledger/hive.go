package syncutils

type MultiMutexLockBuilder struct {
	locks []interface{}
}

func (lockBuilder *MultiMutexLockBuilder) AddLock(identifier interface{}) *MultiMutexLockBuilder {
	lockBuilder.locks = append(lockBuilder.locks, identifier)

	return lockBuilder
}

func (lockBuilder *MultiMutexLockBuilder) Build() []interface{} {
	return lockBuilder.locks
}
