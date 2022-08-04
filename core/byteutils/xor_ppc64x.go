// 3rd-party code from: https://github.com/golang/go/blob/master/src/crypto/cipher/xor_ppc64x.go
//go:build ppc64 || ppc64le
// +build ppc64 ppc64le

package byteutils

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
	_ = dst[n-1]
	xorBytesVSX(&dst[0], &a[0], &b[0], n)
	return n
}

//go:noescape
func xorBytesVSX(dst, a, b *byte, n int)
