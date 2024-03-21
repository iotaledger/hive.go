//go:build stacktrace
// +build stacktrace

//
//nolint:goerr113
package ierrors

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestErrors(t *testing.T) {
	var errWithStacktrace *errorWithStacktrace

	// check that there is no stacktrace included
	err1 := New("err1")
	require.False(t, Is(err1, &errorWithStacktrace{}))

	// check that there is a stacktrace included
	err2 := Join(err1, New("err2"))
	require.ErrorAs(t, err2, &errWithStacktrace)

	err3 := Errorf("err%d", 3)
	require.ErrorAs(t, err3, &errWithStacktrace)

	err4 := Wrap(err1, "err4")
	require.ErrorAs(t, err4, &errWithStacktrace)

	err5 := Wrapf(err1, "%s", "err5")
	require.ErrorAs(t, err5, &errWithStacktrace)

	err6 := WithMessage(err1, "err6")
	require.ErrorAs(t, err6, &errWithStacktrace)

	err7 := WithMessagef(err1, "%s", "err7")
	require.ErrorAs(t, err7, &errWithStacktrace)

	err8 := WithStack(err1)
	require.ErrorAs(t, err8, &errWithStacktrace)

	// check that there is no duplicated stacktrace included
	errStacktrace := WithStack(New("errStacktrace"))
	require.Equal(t, 1, strings.Count(errStacktrace.Error(), "github.com/iotaledger/hive.go/ierrors.TestErrors"))

	err9 := Join(errStacktrace, New("err9"))
	require.Equal(t, 1, strings.Count(err9.Error(), "github.com/iotaledger/hive.go/ierrors.TestErrors"))

	err10 := Errorf("err%d: %w", 10, errStacktrace)
	require.Equal(t, 1, strings.Count(err10.Error(), "github.com/iotaledger/hive.go/ierrors.TestErrors"))

	err11 := Wrap(errStacktrace, "err11")
	require.Equal(t, 1, strings.Count(err11.Error(), "github.com/iotaledger/hive.go/ierrors.TestErrors"))

	err12 := Wrapf(errStacktrace, "%s", "err12")
	require.Equal(t, 1, strings.Count(err12.Error(), "github.com/iotaledger/hive.go/ierrors.TestErrors"))

	err13 := WithMessage(errStacktrace, "err13")
	require.Equal(t, 1, strings.Count(err13.Error(), "github.com/iotaledger/hive.go/ierrors.TestErrors"))

	err14 := WithMessagef(errStacktrace, "%s", "err14")
	require.Equal(t, 1, strings.Count(err14.Error(), "github.com/iotaledger/hive.go/ierrors.TestErrors"))

	err15 := WithStack(errStacktrace)
	require.Equal(t, 1, strings.Count(err15.Error(), "github.com/iotaledger/hive.go/ierrors.TestErrors"))

	err16 := Chain(New("err16"), New("chained"))
	require.Equal(t, 1, strings.Count(err16.Error(), "github.com/iotaledger/hive.go/ierrors.TestErrors"))
}
