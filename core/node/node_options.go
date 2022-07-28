package node

import (
	daemon2 "github.com/iotaledger/hive.go/core/daemon"
)

type NodeOptions struct {
	plugins []*Plugin
	daemon  daemon2.Daemon
}

func newNodeOptions(optionalOptions []NodeOption) *NodeOptions {
	result := &NodeOptions{}

	for _, optionalOption := range optionalOptions {
		optionalOption(result)
	}

	if result.daemon == nil {
		result.daemon = daemon2.New()
	}

	return result
}

type NodeOption func(*NodeOptions)

func Plugins(plugins ...*Plugin) NodeOption {
	return func(args *NodeOptions) {
		args.plugins = append(args.plugins, plugins...)
	}
}

func Daemon(daemon daemon2.Daemon) NodeOption {
	return func(args *NodeOptions) {
		args.daemon = daemon
	}
}
