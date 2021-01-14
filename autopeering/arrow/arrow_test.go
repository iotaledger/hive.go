package arrow

import (
	"github.com/iotaledger/hive.go/identity"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewSalt(t *testing.T) {
	type testCase struct {
		input time.Duration
		want  error
	}

	tests := []testCase{
		{input: 0, want: nil},
		{input: 10, want: nil},
		{input: -1, want: nil},
	}

	for _, test := range tests {
		_, err := NewArRow(test.input, 1, identity.GenerateIdentity(), 100)
		assert.Equal(t, test.want, err, test)
	}
}

func TestSaltExpired(t *testing.T) {
	type testCase struct {
		input time.Duration
		want  bool
	}

	tests := []testCase{
		{input: 0, want: true},
		{input: time.Second * 10, want: false},
		{input: -1, want: true},
	}

	for _, test := range tests {
		arrow, _ := NewArRow(test.input, 1, identity.GenerateIdentity(), 100)
		got := arrow.Expired()
		assert.Equal(t, test.want, got, test)
	}
}

func TestMarshalUnmarshal(t *testing.T) {
	type testCase struct {
		input time.Duration
	}

	tests := []testCase{
		{input: 0},
		{input: time.Second * 10},
		{input: -1},
	}

	for _, test := range tests {
		arrow, _ := NewArRow(test.input, 1, identity.GenerateIdentity(), 100)

		data, err := arrow.Marshal()
		require.Equal(t, nil, err, "NoErrorCheck")

		got, err := Unmarshal(data)
		require.Equal(t, nil, err, "NoErrorCheck")

		assert.Equal(t, arrow.GetRows(), got.GetRows(), "Rows")
		assert.Equal(t, arrow.GetArs(), got.GetArs(), "Ars")
		assert.Equal(t, arrow.GetExpiration().Unix(), got.GetExpiration().Unix(), "SameArrowExpirationTime")

	}
}
