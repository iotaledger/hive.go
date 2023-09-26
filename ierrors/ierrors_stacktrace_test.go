//go:build stacktrace
// +build stacktrace

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
	err2 := Errorf("err%d", 2)
	require.ErrorAs(t, err2, &errWithStacktrace)

	err3 := Wrap(err1, "err3")
	require.ErrorAs(t, err3, &errWithStacktrace)

	err4 := Wrapf(err1, "%s", "err4")
	require.ErrorAs(t, err4, &errWithStacktrace)

	err5 := WithStack(err1)
	require.ErrorAs(t, err5, &errWithStacktrace)

	// check that there is no duplicated stacktrace included
	errStacktrace := WithStack(New("errStacktrace"))
	require.Equal(t, 1, strings.Count(errStacktrace.Error(), "github.com/izuc/zipp.foundation/ierrors.TestErrors"))

	err6 := Errorf("err%d: %w", 6, errStacktrace)
	require.Equal(t, 1, strings.Count(err6.Error(), "github.com/izuc/zipp.foundation/ierrors.TestErrors"))

	err7 := Wrap(errStacktrace, "err7")
	require.Equal(t, 1, strings.Count(err7.Error(), "github.com/izuc/zipp.foundation/ierrors.TestErrors"))

	err8 := Wrapf(errStacktrace, "%s", "err8")
	require.Equal(t, 1, strings.Count(err8.Error(), "github.com/izuc/zipp.foundation/ierrors.TestErrors"))

	err9 := WithStack(errStacktrace)
	require.Equal(t, 1, strings.Count(err9.Error(), "github.com/izuc/zipp.foundation/ierrors.TestErrors"))
}
