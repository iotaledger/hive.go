package bitmask_test

import (
	"testing"

	"github.com/iotaledger/hive.go/bitmask"
)

func TestBitmask(t *testing.T) {
	var b bitmask.BitMask

	if b.HasBit(0) {
		t.Error("flag at pos 0 should not be set")
	}
	if b.HasBit(1) {
		t.Error("flag at pos 1 should not be set")
	}

	b = b.SetBit(0)
	if !b.HasBit(0) {
		t.Error("flag at pos 0 should be set")
	}
	b = b.SetBit(1)
	if !b.HasBit(1) {
		t.Error("flag at pos 1 should be set")
	}

	b = b.ClearBit(0)
	if b.HasBit(0) {
		t.Error("flag at pos 0 should not be set")
	}
	b = b.ClearBit(1)
	if b.HasBit(1) {
		t.Error("flag at pos 1 should not be set")
	}
}
