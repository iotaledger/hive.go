package kzg

import "go.dedis.ch/kyber/v3"

// Commit commits to vector vect[0], ...., vect[D-1]
// it is [f(s)]1 where f is polynomial  in evaluation (Lagrange) form,
// i.e. with f(rou[i]) = vect[i], i = 0..D-1
// vect[k] == nil equivalent to 0
func (sd *TrustedSetup) Commit(vect *[D]kyber.Scalar) kyber.Point {
	ret := sd.Suite.G1().Point().Null()
	elem := sd.Suite.G1().Point()
	for i, e := range vect {
		if e == nil {
			continue
		}
		elem.Mul(e, sd.LagrangeBasis[i])
		ret.Add(ret, elem)
	}
	return ret
}

// Prove returns pi = [(f(s)-vect<index>)/(s-rou<index>)]1
// This isthe proof sent to verifier
func (sd *TrustedSetup) Prove(vect *[D]kyber.Scalar, i int) kyber.Point {
	ret := sd.Suite.G1().Point().Null()
	if vect[i] == nil {
		return ret
	}
	e := sd.Suite.G1().Point()
	for j := range vect {
		if vect[j] == nil {
			continue
		}
		qij := sd.q(vect, i, j)
		e.Mul(qij, sd.LagrangeBasis[j])
		ret.Add(ret, e)
	}
	return ret
}

func (sd *TrustedSetup) q(vect *[D]kyber.Scalar, i, m int) kyber.Scalar {
	ret := sd.Suite.G1().Scalar().Zero()
	if i != m {
		ret.Sub(sd.OmegaPowers[m], sd.OmegaPowers[i])
		ret.Div(vect[m], ret)
	} else {
		e := sd.Suite.G1().Scalar()
		for j := range sd.OmegaPowers {
			if vect[j] == nil || j == m {
				continue
			}
			e.Sub(sd.OmegaPowers[m], sd.OmegaPowers[j])
			e.Div(sd.OmegaPowers[j], e)
			e.Mul(vect[j], e)
			ret.Add(ret, e)
		}
		ret.Div(ret, sd.OmegaPowers[m])
	}
	return ret
}

// Verify verifies KZG proof that polynomial f committed with C has f(rou<atIndex>) = v
// c is commitment to the polynomial
// pi is commitment to the value point (proof)
// value is the value of the polynomial
// adIndex is index of the root of unity where polynomial is expected to have value = v
func (sd *TrustedSetup) Verify(c, pi kyber.Point, v kyber.Scalar, atIndex int) bool {
	p1 := sd.Suite.Pair(pi, sd.Diff2[atIndex])
	e := sd.Suite.G1().Point().Mul(v, nil)
	e.Sub(c, e)
	p2 := sd.Suite.Pair(e, sd.Suite.G2().Point().Base())
	return p1.Equal(p2)
}

// CommitAll return commit to the whole vector and to each of values of it
// Generate commitment to the vector and proofs to all values.
// Expensive. Usually used only in tests
func (sd *TrustedSetup) CommitAll(vect *[D]kyber.Scalar) (kyber.Point, *[D]kyber.Point) {
	retC := sd.Commit(vect)
	retPi := new([D]kyber.Point)
	for i := range vect {
		if vect[i] == nil {
			continue
		}
		retPi[i] = sd.Prove(vect, i)
	}
	return retC, retPi
}
