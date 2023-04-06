package profiling

import (
	"net/http"
	"time"

	// import pprof.
	//nolint:gosec // ToDo: register handlers ourselves.
	_ "net/http/pprof"
	"runtime"

	"github.com/pkg/errors"

	"github.com/iotaledger/hive.go/app"
)

func init() {
	Component = &app.Component{
		Name:   "Profiling",
		Params: params,
		Run:    run,
		IsEnabled: func() bool {
			return ParamsProfiling.Enabled
		},
	}
}

var (
	Component *app.Component
)

func run() error {
	runtime.SetMutexProfileFraction(5)
	runtime.SetBlockProfileRate(5)

	bindAddr := ParamsProfiling.BindAddress

	go func() {
		Component.LogInfof("You can now access the profiling server using: http://%s/debug/pprof/", bindAddr)

		// pprof Server for Debugging
		server := &http.Server{
			Addr:              bindAddr,
			ReadHeaderTimeout: 3 * time.Second,
		}

		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			Component.LogWarnf("Stopped profiling server due to an error (%s)", err)
		}
	}()

	return nil
}
