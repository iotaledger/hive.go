// 3rd-party code from: https://github.com/golang/go/blob/master/src/crypto/cipher/xor_generic.go
// +build !amd64,!ppc64,!ppc64le

package byteutils

import (
	"runtime"
	"unsafe"
)

// XORBytes xors the bytes in a and b. The destination should have enough space, otherwise xorBytes will panic.
// Returns the number of bytes xor'd.
func XORBytes(dst, a, b []byte) int {
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	if n == 0 {
		return 0
	}

	switch {
	case supportsUnaligned:
		fastXORBytes(dst, a, b, n)
	default:
		safeXORBytes(dst, a, b, n)
	}
	return n
}

const wordSize = int(unsafe.Sizeof(uintptr(0)))
const supportsUnaligned = runtime.GOARCH == "386" || runtime.GOARCH == "ppc64" || runtime.GOARCH == "ppc64le" || runtime.GOARCH == "s390x"

// fastXORBytes xors in bulk. It only works on architectures that support unaligned read/writes.
// n needs to be smaller or equal than the length of a and b.
func fastXORBytes(dst, a, b []byte, n int) {
	// Assert dst has enough space
	_ = dst[n-1]

	w := n / wordSize
	if w > 0 {
		dw := *(*[]uintptr)(unsafe.Pointer(&dst))
		aw := *(*[]uintptr)(unsafe.Pointer(&a))
		bw := *(*[]uintptr)(unsafe.Pointer(&b))
		for i := 0; i < w; i++ {
			dw[i] = aw[i] ^ bw[i]
		}
	}

	for i := (n - n%wordSize); i < n; i++ {
		dst[i] = a[i] ^ b[i]
	}
}

// n needs to be smaller or equal than the length of a and b.
func safeXORBytes(dst, a, b []byte, n int) {
	for i := 0; i < n; i++ {
		dst[i] = a[i] ^ b[i]
	}
}
