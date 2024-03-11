package module

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/hive.go/ds"
	"github.com/iotaledger/hive.go/lo"
	"github.com/iotaledger/hive.go/log"
)

func Test(t *testing.T) {
	module1 := New(log.NewLogger(log.WithName("module1")))
	module2 := New(log.NewLogger(log.WithName("module2")))
	module3 := New(log.NewLogger(log.WithName("module3")))

	go func() {
		time.Sleep(2 * time.Second)
		module1.ConstructedEvent().Trigger()

		time.Sleep(2 * time.Second)
		module2.ConstructedEvent().Trigger()

		time.Sleep(2 * time.Second)
		module3.ConstructedEvent().Trigger()
	}()

	wg := WaitAll(Module.ConstructedEvent, module1, module2, module3)

	pendingElements := wg.PendingElements()
	wg.PendingElements().OnUpdate(func(_ ds.SetMutations[Module]) {
		fmt.Println("Pending elements: ", strings.Join(lo.Map(pendingElements.ToSlice(), Module.LogName), ", "))
	})

	wg.Wait()

	require.True(t, module1.ConstructedEvent().WasTriggered())
	require.True(t, module2.ConstructedEvent().WasTriggered())
	require.True(t, module3.ConstructedEvent().WasTriggered())
}
