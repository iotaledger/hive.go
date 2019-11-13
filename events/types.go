package events

func CallbackCaller(handler interface{}, params ...interface{}) {
	handler.(func())()
}

func ErrorCaller(handler interface{}, params ...interface{}) {
	handler.(func(error))(params[0].(error))
}

func IntCaller(handler interface{}, params ...interface{}) {
	handler.(func(int))(params[0].(int))
}

func IntSliceCaller(handler interface{}, params ...interface{}) {
	handler.(func([]int))(params[0].([]int))
}

func Int8Caller(handler interface{}, params ...interface{}) {
	handler.(func(int8))(params[0].(int8))
}

func Int8SliceCaller(handler interface{}, params ...interface{}) {
	handler.(func([]int8))(params[0].([]int8))
}

func Int16Caller(handler interface{}, params ...interface{}) {
	handler.(func(int16))(params[0].(int16))
}

func Int16SliceCaller(handler interface{}, params ...interface{}) {
	handler.(func([]int16))(params[0].([]int16))
}

func Int32Caller(handler interface{}, params ...interface{}) {
	handler.(func(int32))(params[0].(int32))
}

func Int32SliceCaller(handler interface{}, params ...interface{}) {
	handler.(func([]int32))(params[0].([]int32))
}

func Int64Caller(handler interface{}, params ...interface{}) {
	handler.(func(int64))(params[0].(int64))
}

func Int64SliceCaller(handler interface{}, params ...interface{}) {
	handler.(func([]int64))(params[0].([]int64))
}

func Uint8Caller(handler interface{}, params ...interface{}) {
	handler.(func(uint8))(params[0].(uint8))
}

func Uint8SliceCaller(handler interface{}, params ...interface{}) {
	handler.(func([]uint8))(params[0].([]uint8))
}

func Uint16Caller(handler interface{}, params ...interface{}) {
	handler.(func(uint16))(params[0].(uint16))
}

func Uint16SliceCaller(handler interface{}, params ...interface{}) {
	handler.(func([]uint16))(params[0].([]uint16))
}

func Uint32Caller(handler interface{}, params ...interface{}) {
	handler.(func(uint32))(params[0].(uint32))
}

func Uint32SliceCaller(handler interface{}, params ...interface{}) {
	handler.(func([]uint32))(params[0].([]uint32))
}

func Uint64Caller(handler interface{}, params ...interface{}) {
	handler.(func(uint64))(params[0].(uint64))
}

func Uint64SliceCaller(handler interface{}, params ...interface{}) {
	handler.(func([]uint64))(params[0].([]uint64))
}

func ByteCaller(handler interface{}, params ...interface{}) {
	handler.(func(byte))(params[0].(byte))
}

func ByteSliceCaller(handler interface{}, params ...interface{}) {
	handler.(func([]byte))(params[0].([]byte))
}

func StringCaller(handler interface{}, params ...interface{}) {
	handler.(func(string))(params[0].(string))
}

func StringSliceCaller(handler interface{}, params ...interface{}) {
	handler.(func([]string))(params[0].([]string))
}
