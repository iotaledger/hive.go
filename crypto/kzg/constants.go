package kzg

import (
	"math/big"

	"go.dedis.ch/kyber/v3/pairing/bn256"
)

// constants are used in KZG trusted setup

const (
	// factor of order-1
	FACTOR = 16 // 5743
	// D = 257 we will be building 257-ary verkle trie. Each node commits to up to 257 values
	// The indices 0..255 are children, index 256 corresponds to the terminal value if present
	// We will calculate first D values of the Lagrange basis
	D = 16 // 257
	// a constant to check consistency: orderMinus1DivDStr = (fieldOrder-1)/FACTOR
	orderMinus1DivFactorStr = "11318222130532231191502078833773272809084173089657633012342166632255644576"
)

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
	if D > FACTOR {
		panic("D > FACTOR")
	}
	//	c, _ := new(big.Int).SetString(orderMinus1DivFactorStr, 10)
	orderMinus1DivFactor.Sub(fieldOrder, big1)
	orderMinus1DivFactor.Div(orderMinus1DivFactor, bigFactor)
	//	if c.Cmp(orderMinus1DivFactor) != 0 {
	//		panic("inconsistent constants")
	//	}
}
