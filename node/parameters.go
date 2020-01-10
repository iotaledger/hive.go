package node

import (
	flag "github.com/spf13/pflag"
)

const (
	CFG_DISABLE_PLUGINS = "node.disablePlugins"
	CFG_ENABLE_PLUGINS  = "node.enablePlugins"
)

func init() {
	flag.StringSlice(CFG_DISABLE_PLUGINS, nil, "a list of plugins that shall be disabled")
	flag.StringSlice(CFG_ENABLE_PLUGINS, nil, "a list of plugins that shall be enabled")
}
