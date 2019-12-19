package node

import (
	"sync"

	"github.com/iotaledger/hive.go/daemon"
	"github.com/iotaledger/hive.go/logger"
)

var (
	// plugins
	plugins         = make(map[string]int)
	DisabledPlugins = make(map[string]bool)
	EnabledPlugins  = make(map[string]bool)
)

type Node struct {
	wg            *sync.WaitGroup
	loadedPlugins []*Plugin
	Logger        *logger.Logger
}

func New(plugins ...*Plugin) *Node {
	node := &Node{
		wg:            &sync.WaitGroup{},
		loadedPlugins: make([]*Plugin, 0),
		Logger:        logger.NewLogger("Node"),
	}

	// configure the enabled plugins
	node.configure(plugins...)

	return node
}

func Start(plugins ...*Plugin) *Node {
	node := New(plugins...)
	node.Start()

	return node
}

func Run(plugins ...*Plugin) *Node {
	node := New(plugins...)
	node.Run()

	return node
}

func Shutdown() {
	daemon.ShutdownAndWait()
}

func isDisabled(plugin *Plugin) bool {
	_, exists := DisabledPlugins[GetPluginIdentifier(plugin.Name)]

	return exists
}

func isEnabled(plugin *Plugin) bool {
	_, exists := EnabledPlugins[GetPluginIdentifier(plugin.Name)]

	return exists
}

func (node *Node) configure(plugins ...*Plugin) {
	for _, plugin := range plugins {
		status := plugin.Status
		if (status == Enabled && !isDisabled(plugin)) ||
			(status == Disabled && isEnabled(plugin)) {

			plugin.wg = node.wg
			plugin.Node = node

			plugin.Events.Configure.Trigger(plugin)
			node.loadedPlugins = append(node.loadedPlugins, plugin)
			node.Logger.Infof("Loading Plugin: %s ... done", plugin.Name)
		} else {
			node.Logger.Infof("Skipping Plugin: %s", plugin.Name)
		}
	}
}

func (node *Node) Start() {
	node.Logger.Info("Executing plugins...")

	for _, plugin := range node.loadedPlugins {
		plugin.Events.Run.Trigger(plugin)

		node.Logger.Infof("Starting Plugin: %s...", plugin.Name)
	}

	node.Logger.Info("Starting background workers ...")
	daemon.Start()
}

func (node *Node) Run() {
	node.Logger.Info("Executing plugins ...")

	for _, plugin := range node.loadedPlugins {
		plugin.Events.Run.Trigger(plugin)
		node.Logger.Infof("Starting Plugin: %s ... done", plugin.Name)
	}

	node.Logger.Info("Starting background workers ...")

	daemon.Run()

	node.Logger.Info("Shutdown complete!")
}

func AddPlugin(name string, status int) {
	if _, exists := plugins[name]; exists {
		panic("duplicate plugin - \"" + name + "\" was defined already")
	}

	plugins[name] = status

	Events.AddPlugin.Trigger(name, status)
}

func GetPlugins() map[string]int {
	return plugins
}
