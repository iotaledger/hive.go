package profiling

import (
	"github.com/iotaledger/hive.go/app"
)

// ParametersProfiling contains the definition of the parameters used by profiling.
type ParametersProfiling struct {
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
