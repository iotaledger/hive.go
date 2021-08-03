package kzg

import (
	"bytes"
	"golang.org/x/xerrors"
	"io"
	"os"

	"go.dedis.ch/kyber/v3"
	"go.dedis.ch/kyber/v3/pairing/bn256"
)

// TrustedSetup is a trusted setup for KZG calculations with degree D.
// It includes roots of unity for the scalar field and values on G1 and G2 curves which are pre-calculated
// from the secret. The secret itself must be destroyed immediately after trusted setup is generated.
// The trusted setup is a public value. Normally it is stored as a public constant or file and is loaded from there.
// It is impossible to restore secret from the trusted setup
// [x]1 means a projection of scalar x to the G1 curve. [x]1 = xG, where G is the generating element
// [x]2 means a projection of scalar x to the G2 curve. [x]2 = xH, where H is the generating element
type TrustedSetup struct {
	Suite               *bn256.Suite
	RootOfUnity         kyber.Scalar      // rou: a primitive root of unity of the field
	RootOfUnityPowers   [D]kyber.Scalar   // rou<i> =  rou^i. rou<0> == 1, row<1> =  rou
	Diff2               [D]kyber.Point    // [secret-rou<i>]2
	DiffInv             [D]kyber.Point    // [1/(s-rou<i>)]1
	LagrangePolysCommit [D]kyber.Point    // [l<i>(secret)]1
	PiMatrix            [D][D]kyber.Point // [l<i>(s)/(s-rou<j>)]1
	// util constants
	ZeroG1 kyber.Scalar
	OneG1  kyber.Scalar
	DG1    kyber.Scalar
}

var (
	errWrongSecret     = xerrors.New("wrong secret")
	errNotROU          = xerrors.New("not a root of unity")
	errNonPrimitiveROU = xerrors.New("root of unity must be primitive")
)

func newTrustedSetup(suite *bn256.Suite) *TrustedSetup {
	ret := &TrustedSetup{
		Suite:       suite,
		RootOfUnity: suite.G1().Scalar(),
		// util
		ZeroG1: suite.G1().Scalar().Zero(),
		OneG1:  suite.G1().Scalar().One(),
		DG1:    suite.G1().Scalar().SetInt64(D),
	}
	for i := range ret.RootOfUnityPowers {
		ret.RootOfUnityPowers[i] = suite.G1().Scalar()
	}
	for i := 0; i < D; i++ {
		ret.LagrangePolysCommit[i] = suite.G1().Point()
		ret.DiffInv[i] = suite.G1().Point()
		ret.Diff2[i] = suite.G2().Point()
		for j := 0; j < D; j++ {
			ret.PiMatrix[i][j] = suite.G1().Point()
		}
	}
	return ret
}

// TrustedSetupFromSecret calculates TrustedSetup from secret and rootOfUnity
// Only used once after what secret must be destroyed
// The trusted setup does not contain any secret
func TrustedSetupFromSecret(suite *bn256.Suite, rootOfUnity, secret kyber.Scalar) (*TrustedSetup, error) {
	ret := newTrustedSetup(suite)
	if err := ret.generate(rootOfUnity, secret); err != nil {
		return nil, err
	}
	return ret, nil
}

// TrustedSetupFromBytes unmarshals trusted setup from binary representation
func TrustedSetupFromBytes(suite *bn256.Suite, data []byte) (*TrustedSetup, error) {
	ret := newTrustedSetup(suite)
	if err := ret.read(bytes.NewReader(data)); err != nil {
		return nil, err
	}
	return ret, nil
}

// TrustedSetupFromFile restores trusted setup from file
func TrustedSetupFromFile(suite *bn256.Suite, fname string) (*TrustedSetup, error) {
	data, err := os.ReadFile(fname)
	if err != nil {
		return nil, err
	}
	return TrustedSetupFromBytes(suite, data)
}

func (sd *TrustedSetup) Bytes() []byte {
	var buf bytes.Buffer
	if err := sd.write(&buf); err != nil {
		panic(err)
	}
	return buf.Bytes()
}

// generate fills up the TrustedSetup based on its root of unity and provided secret
func (sd *TrustedSetup) generate(rootOfUnity, secret kyber.Scalar) error {
	sd.RootOfUnity.Set(rootOfUnity)
	// check if secret is not equal to any power of root of unity
	for i := range sd.RootOfUnityPowers {
		if secret.Equal(sd.RootOfUnityPowers[i]) {
			return errWrongSecret
		}
		powerSimple(sd.Suite, sd.RootOfUnity, i, sd.RootOfUnityPowers[i])
		if i > 0 && sd.RootOfUnityPowers[i].Equal(sd.OneG1) {
			return errNonPrimitiveROU
		}
	}
	// generate [secret-rou<i>]2
	e2 := sd.Suite.G2().Scalar()
	for i := range sd.Diff2 {
		e2.Sub(secret, sd.RootOfUnityPowers[i])
		sd.Diff2[i].Mul(e2, nil)
	}
	// Pre-calculate auxiliary values on the curve
	e1 := sd.Suite.G1().Scalar()
	for i := range sd.DiffInv {
		e1.Sub(secret, sd.RootOfUnityPowers[i])
		e1.Inv(e1)
		sd.DiffInv[i].Mul(e1, nil) // = [1/(secret-rou<i>)]1
	}

	for i := range sd.LagrangePolysCommit {
		l := sd.evalLagrangeValue(i, secret)
		sd.LagrangePolysCommit[i].Mul(l, nil) // [l<i>(secret)]1
		for j := range sd.PiMatrix[i] {
			e1.Sub(secret, sd.RootOfUnityPowers[j])
			e1.Inv(e1)
			e1.Mul(l, e1)
			sd.PiMatrix[i][j].Mul(e1, nil)
		}
	}
	return nil
}

// evalLagrangeValue calculates li(X) = [prod<j=0,D-1;j!=i>((X-rou^j)/(rou^i-rou^j)]1
func (sd *TrustedSetup) evalLagrangeValue(i int, v kyber.Scalar) kyber.Scalar {
	ret := sd.Suite.G1().Scalar().One()
	numer := sd.Suite.G1().Scalar()
	denom := sd.Suite.G1().Scalar()
	elem := sd.Suite.G1().Scalar()
	for j := 0; j < D; j++ {
		if j == i {
			continue
		}
		numer.Sub(v, sd.RootOfUnityPowers[j])
		denom.Sub(sd.RootOfUnityPowers[i], sd.RootOfUnityPowers[j])
		elem.Div(numer, denom)
		ret.Mul(ret, elem)
	}
	return ret
}

func (sd *TrustedSetup) write(w io.Writer) error {
	if _, err := sd.RootOfUnity.MarshalTo(w); err != nil {
		return err
	}
	for i := range sd.Diff2 {
		if _, err := sd.Diff2[i].MarshalTo(w); err != nil {
			return err
		}
	}
	for i := range sd.LagrangePolysCommit {
		if _, err := sd.LagrangePolysCommit[i].MarshalTo(w); err != nil {
			return err
		}
	}
	for i := range sd.DiffInv {
		if _, err := sd.DiffInv[i].MarshalTo(w); err != nil {
			return err
		}
	}
	for i := range sd.PiMatrix {
		for j := range sd.PiMatrix[i] {
			if _, err := sd.PiMatrix[i][j].MarshalTo(w); err != nil {
				return err
			}
		}
	}
	return nil
}

func (sd *TrustedSetup) read(r io.Reader) error {
	if _, err := sd.RootOfUnity.UnmarshalFrom(r); err != nil {
		return err
	}
	if !isRootOfUnity(sd.Suite, sd.RootOfUnity) {
		return errNotROU
	}
	// calculate powers
	for i := range sd.RootOfUnityPowers {
		powerSimple(sd.Suite, sd.RootOfUnity, i, sd.RootOfUnityPowers[i])
		if i > 0 && sd.RootOfUnityPowers[i].Equal(sd.OneG1) {
			// should not be 1
			return errNonPrimitiveROU
		}
	}
	for i := range sd.Diff2 {
		if _, err := sd.Diff2[i].UnmarshalFrom(r); err != nil {
			return err
		}
	}
	for i := range sd.LagrangePolysCommit {
		if _, err := sd.LagrangePolysCommit[i].UnmarshalFrom(r); err != nil {
			return err
		}
	}
	for i := range sd.DiffInv {
		if _, err := sd.DiffInv[i].UnmarshalFrom(r); err != nil {
			return err
		}
	}
	for i := range sd.PiMatrix {
		for j := range sd.PiMatrix[i] {
			if _, err := sd.PiMatrix[i][j].UnmarshalFrom(r); err != nil {
				return err
			}
		}
	}
	return nil
}
