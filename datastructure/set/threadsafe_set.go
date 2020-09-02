package set

type threadSafeSet struct{}

func newThreadSafeSet() *threadSafeSet {
	return &threadSafeSet{}
}

func (set *threadSafeSet) Add(element interface{}) bool {
	panic("implement me")
}

func (set *threadSafeSet) Delete(element interface{}) bool {
	panic("implement me")
}

func (set *threadSafeSet) Has(element interface{}) bool {
	panic("implement me")
}

func (set *threadSafeSet) ForEach(callback func(element interface{})) {
	panic("implement me")
}

func (set *threadSafeSet) Clear() {
	panic("implement me")
}

func (set *threadSafeSet) Size() int {
	panic("implement me")
}

// code contract - make sure the type implements the interface
var _ Set = &threadSafeSet{}
