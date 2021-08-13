package kzg

import (
	"encoding/hex"
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"
	"go.dedis.ch/kyber/v3"
	"go.dedis.ch/kyber/v3/pairing/bn256"
	"go.dedis.ch/kyber/v3/util/random"
	"golang.org/x/crypto/blake2b"
)

func TestConst(t *testing.T) {
	t.Logf("FACTOR = %d", FACTOR)
	t.Logf("D = %d", D)
	t.Logf("fieldOrder = %d", fieldOrder)
	orderMinus1 := new(big.Int)
	orderMinus1.Sub(fieldOrder, big1)
	orderMinus1DivFactor.Div(orderMinus1, bigFactor)
	t.Logf("(fieldOrder-1)/FACTOR = %d", orderMinus1DivFactor)
	mod := new(big.Int)
	mod.Mod(orderMinus1, bigFactor)
	t.Logf("(fieldOrder-1)%%FACTOR = %d", mod)

	suite := bn256.NewSuite()
	t.Logf("G1().Scalarlen: %d", suite.G1().ScalarLen())
	t.Logf("G1().Pointllen: %d", suite.G1().PointLen())
	t.Logf("G2().Scalarlen: %d", suite.G2().ScalarLen())
	t.Logf("G2().Pointllen: %d", suite.G2().PointLen())
	t.Logf("GT().Scalarlen: %d", suite.GT().ScalarLen())
	t.Logf("GT().Pointllen: %d", suite.GT().PointLen())
}

func TestGenerate(t *testing.T) {
	suite := bn256.NewSuite()
	rou, _ := GenRootOfUnityQuasiPrimitive(suite)
	t.Logf("omega = %s", rou.String())
	secret := suite.G1().Scalar().Pick(random.New())
	tr, err := TrustedSetupFromSecret(suite, rou, secret)
	require.NoError(t, err)
	data := tr.Bytes()
	t.Logf("trusted setup size: %d", len(data))

	trBack, err := TrustedSetupFromBytes(suite, data)
	require.NoError(t, err)

	require.EqualValues(t, tr.Bytes(), trBack.Bytes())
	h := blake2b.Sum256(data)
	t.Logf("hash = %s", hex.EncodeToString(h[:]))
}

func TestValidate0(t *testing.T) {
	suite := bn256.NewSuite()
	rou, _ := GenRootOfUnityQuasiPrimitive(suite)
	t.Logf("omega = %s", rou.String())
	secret := suite.G1().Scalar().Pick(random.New())
	tr, err := TrustedSetupFromSecret(suite, rou, secret)
	require.NoError(t, err)

	vect := new([D]kyber.Scalar)
	vect[0] = tr.Suite.G1().Scalar().SetInt64(42)
	c := tr.Commit(vect)

	t.Logf("C = %s", c)
	pi0 := tr.Prove(vect, 0)
	pi1 := tr.Prove(vect, 1)
	t.Logf("Pi[0] = %s", pi0)
	t.Logf("Pi[1] = %s", pi1)

	require.True(t, tr.Verify(c, pi0, vect[0], 0))
	require.True(t, tr.Verify(c, pi1, tr.ZeroG1, 1))
}

func TestValidate1(t *testing.T) {
	suite := bn256.NewSuite()
	rou, _ := GenRootOfUnityQuasiPrimitive(suite)
	t.Logf("omega = %s", rou.String())
	secret := suite.G1().Scalar().Pick(random.New())
	tr, err := TrustedSetupFromSecret(suite, rou, secret)
	require.NoError(t, err)

	vect := new([D]kyber.Scalar)
	vect[0] = tr.Suite.G1().Scalar().SetInt64(42)
	c := tr.Commit(vect)
	t.Logf("C = %s", c)
	c, pi := tr.CommitAll(vect)
	require.True(t, tr.Verify(c, pi[0], vect[0], 0))
}

func TestValidate2(t *testing.T) {
	suite := bn256.NewSuite()
	rou, _ := GenRootOfUnityQuasiPrimitive(suite)
	secret := suite.G1().Scalar().Pick(random.New())
	tr, err := TrustedSetupFromSecret(suite, rou, secret)
	require.NoError(t, err)

	vect := new([D]kyber.Scalar)
	for i := range vect {
		vect[i] = tr.Suite.G1().Scalar().SetInt64(int64(i))
	}
	c := tr.Commit(vect)
	t.Logf("C = %s", c)
	c, pi := tr.CommitAll(vect)
	for i := range vect {
		require.True(t, tr.Verify(c, pi[i], vect[i], i))
	}
	v := tr.Suite.G1().Scalar()
	for i := range vect {
		v.SetInt64(int64(i + 1))
		require.False(t, tr.Verify(c, pi[i], v, i))
	}
	rnd := random.New()
	for k := 0; k < 100; k++ {
		v.Pick(rnd)
		for i := range vect {
			require.False(t, tr.Verify(c, pi[i], v, i))
		}
	}
}

func TestValidate1Load(t *testing.T) {
	suite := bn256.NewSuite()
	tr, err := TrustedSetupFromFile(suite, "example.setup")
	require.NoError(t, err)

	vect := new([D]kyber.Scalar)
	vect[0] = tr.Suite.G1().Scalar().SetInt64(42)
	c := tr.Commit(vect)
	t.Logf("C = %s", c)
	c, pi := tr.CommitAll(vect)
	require.True(t, tr.Verify(c, pi[0], vect[0], 0))
}

func TestValidate2Load(t *testing.T) {
	suite := bn256.NewSuite()
	tr, err := TrustedSetupFromFile(suite, "example.setup")
	require.NoError(t, err)

	vect := new([D]kyber.Scalar)
	for i := range vect {
		vect[i] = tr.Suite.G1().Scalar().SetInt64(int64(i))
	}
	c := tr.Commit(vect)
	t.Logf("C = %s", c)
	c, pi := tr.CommitAll(vect)
	for i := range vect {
		require.True(t, tr.Verify(c, pi[i], vect[i], i))
	}
	v := tr.Suite.G1().Scalar()
	for i := range vect {
		v.SetInt64(int64(i + 1))
		require.False(t, tr.Verify(c, pi[i], v, i))
	}
	rnd := random.New()
	for k := 0; k < 100; k++ {
		v.Pick(rnd)
		for i := range vect {
			require.False(t, tr.Verify(c, pi[i], v, i))
		}
	}
}
