package node

import (
	"github.com/iotaledger/hive.go/logger"
	flag "github.com/spf13/pflag"
)

const (
	CFG_LOG_LEVEL       = "node.LogLevel"
	CFG_DISABLE_PLUGINS = "node.disablePlugins"
	CFG_ENABLE_PLUGINS  = "node.enablePlugins"
)

func init() {
	flag.Int(CFG_LOG_LEVEL, int(logger.LevelInfo), "controls the log types that are shown")
	flag.String(CFG_DISABLE_PLUGINS, "", "a list of plugins that shall be disabled")
	flag.String(CFG_ENABLE_PLUGINS, "", "a list of plugins that shall be enabled")
}
