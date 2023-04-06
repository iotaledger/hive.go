package profiling

import (
	"github.com/iotaledger/hive.go/app"
)

// ParametersProfiling contains the definition of the parameters used by profiling.
type ParametersProfiling struct {
	// Enabled defines whether the profiling component is enabled.
	Enabled bool `default:"false" usage:"whether the profiling component is enabled"`
	// the bind address on which the profiler listens on
	BindAddress string `default:"localhost:6060" usage:"the bind address on which the profiler listens on"`
}

var ParamsProfiling = &ParametersProfiling{}

var params = &app.ComponentParams{
	Params: map[string]any{
		"profiling": ParamsProfiling,
	},
	Masked: nil,
}
