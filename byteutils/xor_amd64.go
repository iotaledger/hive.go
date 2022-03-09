// 3rd-party code from: https://github.com/golang/go/blob/master/src/crypto/cipher/xor_amd64.go
//go:build amd64
// +build amd64

package byteutils

// XORBytes xors the bytes in a and b. The destination should have enough space, otherwise XOR will panic.
// Returns the number of bytes xor'd.
func XORBytes(dst, a, b []byte) int {
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	if n == 0 {
		return 0
	}
	_ = dst[n-1]
	xorBytesSSE2(&dst[0], &a[0], &b[0], n) // amd64 must have SSE2
	return n
}

//go:noescape
func xorBytesSSE2(dst, a, b *byte, n int)
