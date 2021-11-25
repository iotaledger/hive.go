package node_test

import (
	"testing"

	"github.com/iotaledger/hive.go/v2/configuration"
	"github.com/iotaledger/hive.go/v2/events"
	"github.com/iotaledger/hive.go/v2/logger"
	"github.com/iotaledger/hive.go/v2/node"
	"github.com/stretchr/testify/require"
	"go.uber.org/dig"
)

func TestDependencyInjection(t *testing.T) {
	type PluginADeps struct {
		dig.In
		DepFromB string `name:"providedByB"`
	}

	stringVal := "到月球"

	depsA := &PluginADeps{}
	pluginA := node.NewPlugin("A", depsA, node.Enabled,
		func(plugin *node.Plugin) {
			require.Equal(t, stringVal, depsA.DepFromB)
		},
	)

	pluginB := node.NewPlugin("B", nil, node.Enabled)

	pluginB.Events.Init.Attach(events.NewClosure(func(_ *node.Plugin, container *dig.Container) {
		require.NoError(t, container.Provide(func() string {
			return stringVal
		}, dig.Name("providedByB")))
	}))

	require.NoError(t, logger.InitGlobalLogger(configuration.New()))
	node.Run(node.Plugins(pluginA, pluginB))
}
