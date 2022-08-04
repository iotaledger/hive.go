package debug

import (
	"fmt"
	"testing"
)

func TestNewContext(t *testing.T) {
	callerContext := CallerStackTrace()

	closureContext := ClosureStackTrace(func() {
		fmt.Println(callerContext)
	})

	fmt.Println(callerContext)
	fmt.Println(closureContext)
}
