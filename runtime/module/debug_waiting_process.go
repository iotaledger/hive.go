package module

import (
	"fmt"
	"strings"
	"sync/atomic"
	"time"

	"github.com/iotaledger/hive.go/ds"
	"github.com/iotaledger/hive.go/lo"
)

func DebugWaitingProcess(reportInterval time.Duration, handler ...func(ds.Set[Module])) {
	debugReportInterval.Store(&reportInterval)

	if len(handler) > 0 {
		debugReportHandler.Store(&handler[0])
	} else {
		defaultDebugWaitAllHandler := func(pendingModules ds.Set[Module]) {
			fmt.Println("Waiting for: " + strings.Join(lo.Map(pendingModules.ToSlice(), Module.LogName), ", "))
		}

		debugReportHandler.Store(&defaultDebugWaitAllHandler)
	}
}

func reportPendingModules(modules ...Module) (pendingModules ds.Set[Module]) {
	reportInterval := debugReportInterval.Load()
	reportHandler := debugReportHandler.Load()
	if reportInterval == nil || reportHandler == nil {
		return
	}

	pendingModules = ds.NewSet[Module]()
	for _, module := range modules {
		pendingModules.Add(module)
	}

	go func() {
		ticker := time.NewTicker(*reportInterval)
		defer ticker.Stop()

		for range ticker.C {
			if pendingModules.IsEmpty() {
				break
			}

			(*reportHandler)(pendingModules)
		}
	}()

	return
}

var (
	debugReportInterval atomic.Pointer[time.Duration]
	debugReportHandler  atomic.Pointer[func(ds.Set[Module])]
)
