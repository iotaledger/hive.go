package kzg

import (
	"math/big"

	"go.dedis.ch/kyber/v3/pairing/bn256"
)

// constants are used in KZG trusted setup

// factor of order-1
const FACTOR = 5743

var (
	// from kyber library
	// Order is the number of elements in both G₁ and G₂: 36u⁴+36u³+18u²+6u+1.
	// order-1 = (2**5) * 3 * 5743 * 280941149 * 130979359433191 * 491513138693455212421542731357 * 6518589491078791937
	fieldOrder = bn256.Order
	// orderMinus1DivFactor used in calculation of roots of unity. (fieldOrder-1)/16 == orderMinus1DivDStr
	orderMinus1DivFactor = new(big.Int)
	big0                 = new(big.Int).SetInt64(0)
	big1                 = new(big.Int).SetInt64(1)
	big2                 = new(big.Int).SetInt64(2)
	bigFactor            = new(big.Int).SetInt64(FACTOR)
	//bigD                 = new(big.Int).SetInt64(D)
)

// check consistency of constants
func init() {
	orderMinus1DivFactor.Sub(fieldOrder, big1)
	orderMinus1DivFactor.Div(orderMinus1DivFactor, bigFactor)
	orderMinus1ModFactor := new(big.Int)
	if orderMinus1ModFactor.Cmp(big0) != 0 {
		panic("inconsistent constants")
	}
}
