package kzg

import (
	"math/big"

	"go.dedis.ch/kyber/v3/pairing/bn256"
)

// constants are used in KZG trusted setup

const (
	// D = 2**LOGD = 16 is an optimal factor of (fieldOrder-1)
	D = 2 * 2 * 2 * 2
	// LOGD just = 4 to remember we only need 4 bits for the index in the vector elements
	LOGD = 4
	// a constant to check consistency: orderMinus1DivDStr = (fieldOrder-1)/D
	orderMinus1DivDStr = "4062534355977912733299777421397494108910650378368986649367566435565260424998"
)

var (
	// from kyber library
	// Order is the number of elements in both G₁ and G₂: 36u⁴+36u³+18u²+6u+1.
	// order-1 = (2**5) * 3 * 5743 * 280941149 * 130979359433191 * 491513138693455212421542731357 * 6518589491078791937
	fieldOrder = bn256.Order
	// orderMinus1DivD used in calculation of roots of unity. (fieldOrder-1)/16 == orderMinus1DivDStr
	orderMinus1DivD = new(big.Int)
	big0            = new(big.Int).SetInt64(0)
	big1            = new(big.Int).SetInt64(1)
	big2            = new(big.Int).SetInt64(2)
	bigD            = new(big.Int).SetInt64(D)
)

// check consistency of constants
func init() {
	c, _ := new(big.Int).SetString(orderMinus1DivDStr, 10)
	orderMinus1DivD.Sub(fieldOrder, big1)
	orderMinus1DivD.Div(orderMinus1DivD, bigD)
	if c.Cmp(orderMinus1DivD) != 0 {
		panic("inconsistent constants")
	}
}
