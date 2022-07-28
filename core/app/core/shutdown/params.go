package shutdown

import (
	"time"

	"github.com/iotaledger/hive.go/core/app"
)

// ParametersShutdown contains the definition of the parameters used by shutdown.
type ParametersShutdown struct {
	// the maximum time to wait for background processes to finish during shutdown before terminating the app.
	StopGracePeriod time.Duration `default:"300s"  usage:"the maximum time to wait for background processes to finish during shutdown before terminating the app"`
}

var ParamsShutdown = &ParametersShutdown{}

var params = &app.ComponentParams{
	Params: map[string]any{
		"app": ParamsShutdown,
	},
	Masked: nil,
}
