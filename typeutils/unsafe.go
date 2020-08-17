package typeutils

import (
	"reflect"
	"runtime"
	"unsafe"
)

// Converts a slice of bytes into a string without performing a copy.
// NOTE: This is an unsafe operation and may lead to problems if the bytes
// passed as argument are changed while the string is used.  No checking whether
// bytes are valid UTF-8 data is performed.
func BytesToString(b []byte) string {
	// ensure the underlying bytes don't get GC'ed before the assignment happens
	defer runtime.KeepAlive(&b)

	return *(*string)(unsafe.Pointer(&b))
}

// Converts a string into a slice of bytes without performing a copy.
// NOTE: This is an unsafe operation and may lead to problems if the bytes are changed.
func StringToBytes(s string) []byte {
	// ensure the underlying string doesn't get GC'ed before the assignment happens
	defer runtime.KeepAlive(&s)

	sh := (*reflect.StringHeader)(unsafe.Pointer(&s))
	b := *(*[]byte)(unsafe.Pointer(&reflect.SliceHeader{
		Data: sh.Data,
		Len:  sh.Len,
		Cap:  sh.Len,
	}))

	return b
}
