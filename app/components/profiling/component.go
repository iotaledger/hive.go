package profiling

import (
	"net/http"
	//nolint:gosec // ToDo: register handlers ourselves.
	_ "net/http/pprof"
	"runtime"
	"time"

	"github.com/felixge/fgprof"
	"go.uber.org/dig"

	"github.com/iotaledger/hive.go/app"
	"github.com/iotaledger/hive.go/ierrors"
)

func init() {
	Component = &app.Component{
		Name:   "Profiling",
		Params: params,
		IsEnabled: func(_ *dig.Container) bool {
			return ParamsProfiling.Enabled
		},
		Run: run,
	}
}

var (
	Component *app.Component
)

func run() error {
	runtime.SetMutexProfileFraction(5)
	runtime.SetBlockProfileRate(5)

	bindAddr := ParamsProfiling.BindAddress

	http.DefaultServeMux.Handle("/debug/fgprof", fgprof.Handler())

	go func() {
		Component.LogInfof("You can now access the profiling server using: http://%s/debug/pprof/", bindAddr)

		// pprof Server for Debugging
		server := &http.Server{
			Addr:              bindAddr,
			ReadHeaderTimeout: 3 * time.Second,
		}

		if err := server.ListenAndServe(); err != nil && !ierrors.Is(err, http.ErrServerClosed) {
			Component.LogWarnf("Stopped profiling server due to an error (%s)", err)
		}
	}()

	return nil
}
