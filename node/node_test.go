package node_test

import (
	"testing"

	"github.com/iotaledger/hive.go/configuration"
	"github.com/iotaledger/hive.go/events"
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/hive.go/node"
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
	pluginA := node.NewPlugin("A", node.Enabled, depsA,
		func(plugin *node.Plugin) {
			require.Equal(t, stringVal, depsA.DepFromB)
		},
	)

	pluginB := node.NewPlugin("B", node.Enabled, depsA,
		func(plugin *node.Plugin) {

		},
	)

	pluginB.Events.Init.Attach(events.NewClosure(func(_ *node.Plugin, container *dig.Container) {
		require.NoError(t, container.Provide(func() string {
			return stringVal
		}, dig.Name("providedByB")))
	}))

	require.NoError(t, logger.InitGlobalLogger(configuration.New()))
	node.Run(node.Plugins(pluginA, pluginB))
}
