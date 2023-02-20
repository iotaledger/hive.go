package syncutils

type KRWMutex struct {
	keyMutexConsumers map[interface{}]int
	keyMutexes        map[interface{}]*RWMutex
	mutex             RWMutex
}

func NewKRWMutex() *KRWMutex {
	return &KRWMutex{
		keyMutexConsumers: make(map[interface{}]int),
		keyMutexes:        make(map[interface{}]*RWMutex),
	}
}

func (k *KRWMutex) Register(key interface{}) (result *RWMutex) {
	k.mutex.Lock()

	if val, exists := k.keyMutexConsumers[key]; exists {
		k.keyMutexConsumers[key] = val + 1
		result = k.keyMutexes[key]
	} else {
		result = &RWMutex{}

		k.keyMutexConsumers[key] = 1
		k.keyMutexes[key] = result
	}

	k.mutex.Unlock()

	return
}

func (k *KRWMutex) Free(key interface{}) {
	k.mutex.Lock()

	if val, exists := k.keyMutexConsumers[key]; exists {
		if val == 1 {
			delete(k.keyMutexConsumers, key)
			delete(k.keyMutexes, key)
		} else {
			k.keyMutexConsumers[key] = val - 1
		}
	} else {
		panic("trying to free non-existent key")
	}

	k.mutex.Unlock()
}
