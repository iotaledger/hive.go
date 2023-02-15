package shutdown

import (
	"time"

	"github.com/iotaledger/hive.go/app"
)

// ParametersShutdown contains the definition of the parameters used by shutdown.
type ParametersShutdown struct {
	// the maximum time to wait for background processes to finish during shutdown before terminating the app.
	StopGracePeriod time.Duration `default:"300s" usage:"the maximum time to wait for background processes to finish during shutdown before terminating the app"`

	Log struct {
		// whether to store self-shutdown events to a log file.
		Enabled bool `default:"true" usage:"whether to store self-shutdown events to a log file"`
		// the file path to the self-shutdown log.
		FilePath string `default:"shutdown.log" usage:"the file path to the self-shutdown log"`
	}
}

var ParamsShutdown = &ParametersShutdown{}

var params = &app.ComponentParams{
	Params: map[string]any{
		"app.shutdown": ParamsShutdown,
	},
	Masked: nil,
}
