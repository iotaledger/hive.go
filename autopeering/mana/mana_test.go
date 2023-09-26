package mana

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/izuc/zipp.foundation/crypto/ed25519"
	"github.com/izuc/zipp.foundation/crypto/identity"
)

// 100
// 20
// 10 10
// 4 4 4 4
// 3 3 3  3
// 2 2 2 2 2 2
// 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1

func newTestIdentity(name string) *identity.Identity {
	key := ed25519.PublicKey{}
	copy(key[:], name)

	return identity.New(key)
}

func newTestIdentities(n int) (identities []*identity.Identity) {
	identities = make([]*identity.Identity, n)
	for i := 0; i < n; i++ {
		identities[i] = newTestIdentity(fmt.Sprintf("%d", i))
	}

	return
}

var testIdentityMana map[*identity.Identity]uint64

func newTestMana(identities []*identity.Identity) (m map[*identity.Identity]uint64) {
	m = make(map[*identity.Identity]uint64, len(identities))
	for i, p := range identities {
		m[p] = uint64(i)
	}

	return m
}

type manaFunc Func

var mana manaFunc

func (f manaFunc) Eval(identity *identity.Identity) uint64 {
	return testIdentityMana[identity]
}

func TestRankByFixedRange(t *testing.T) {
	IDs := newTestIdentities(10)
	// 0 1 2 3 4 5 6 7 8 9
	testIdentityMana = newTestMana(IDs)
	R := 2

	set := RankByFixedRange(mana.Eval, IDs[0], IDs, R)
	expectedSet := IDs[1:3]
	require.ElementsMatch(t, expectedSet, set)

	set = RankByFixedRange(mana.Eval, IDs[len(IDs)-1], IDs, R)
	expectedSet = IDs[len(IDs)-1-R : len(IDs)-1]
	require.ElementsMatch(t, expectedSet, set)

	set = RankByFixedRange(mana.Eval, IDs[len(IDs)/2], IDs, R)
	expectedSet = IDs[(len(IDs)/2)-R : (len(IDs) / 2)]
	expectedSet = append(expectedSet, IDs[(len(IDs)/2)+1:(len(IDs)/2)+R+1]...)
	require.ElementsMatch(t, expectedSet, set)
}

func TestRankByVariableRange(t *testing.T) {
	IDs := newTestIdentities(10)
	// 0 1 2 3 4 5 6 7 8 9
	testIdentityMana = newTestMana(IDs)
	R := 2
	ro := 1.51

	set := RankByVariableRange(mana.Eval, IDs[0], IDs, R, ro)
	expectedSet := IDs[1:3]
	require.ElementsMatch(t, expectedSet, set)

	set = RankByVariableRange(mana.Eval, IDs[len(IDs)-1], IDs, R, ro)
	expectedSet = IDs[len(IDs)-2-R : len(IDs)-1]
	require.ElementsMatch(t, expectedSet, set)

	set = RankByVariableRange(mana.Eval, IDs[len(IDs)/2], IDs, R, ro)
	expectedSet = IDs[(len(IDs)/2)-R : (len(IDs) / 2)]
	expectedSet = append(expectedSet, IDs[(len(IDs)/2)+1:(len(IDs)/2)+R+1]...)
	require.ElementsMatch(t, expectedSet, set)
}

func TestRankByThreshold(t *testing.T) {
	IDs := newTestIdentities(10)
	testIdentityMana = newTestMana(IDs)
	threshold := 1. / 3.

	expectedMana := uint64(float64(Total(mana.Eval, IDs)) * threshold)

	set := RankByThreshold(mana.Eval, IDs[0], IDs, threshold)
	actualMana := Total(mana.Eval, set)
	require.GreaterOrEqual(t, actualMana, expectedMana)

	set = RankByThreshold(mana.Eval, IDs[len(IDs)-1], IDs, threshold)
	actualMana = Total(mana.Eval, set)
	require.GreaterOrEqual(t, actualMana, expectedMana)

	set = RankByThreshold(mana.Eval, IDs[len(IDs)/2], IDs, threshold)
	actualMana = Total(mana.Eval, set)
	require.GreaterOrEqual(t, actualMana, expectedMana)
}
