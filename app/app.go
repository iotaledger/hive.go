package app

import (
	"context"
	"fmt"
	"math"
	"os"
	"sort"
	"strings"
	"time"

	goversion "github.com/hashicorp/go-version"
	flag "github.com/spf13/pflag"
	"go.uber.org/dig"

	"github.com/iotaledger/hive.go/configuration"
	"github.com/iotaledger/hive.go/daemon"
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/hive.go/timeutil"
	"github.com/iotaledger/hive.go/version"
)

const (
	DefaultFlagSetName = "appConfig"
)

const (
	// CfgAppCheckForUpdates defines whether to check for updates of the application or not.
	CfgAppCheckForUpdates = "app.checkForUpdates"
	// CfgAppDisablePlugins defines a list of plugins that shall be disabled.
	CfgAppDisablePlugins = "app.disablePlugins"
	// CfgAppEnablePlugins defines a list of plugins that shall be enabled.
	CfgAppEnablePlugins = "app.enablePlugins"
)

// AppInfo provides informations about the app.
type AppInfo struct {
	Name                string
	Version             string
	LatestGitHubVersion string
}

type App struct {
	appInfo                 *AppInfo
	enabledPlugins          map[string]struct{}
	disabledPlugins         map[string]struct{}
	forceDisabledComponents map[string]struct{}
	coreComponentsMap       map[string]*CoreComponent
	coreComponents          []*CoreComponent
	pluginsMap              map[string]*Plugin
	plugins                 []*Plugin
	container               *dig.Container
	log                     *logger.Logger
	appFlagSet              *flag.FlagSet
	appConfig               *configuration.Configuration
	configs                 ConfigurationSets
	maskedKeys              []string
	options                 *AppOptions
}

func New(name string, version string, optionalOptions ...AppOption) *App {
	appOpts := &AppOptions{}
	appOpts.apply(defaultAppOptions...)
	appOpts.apply(optionalOptions...)

	a := &App{
		appInfo: &AppInfo{
			Name:                name,
			Version:             version,
			LatestGitHubVersion: "",
		},
		enabledPlugins:          make(map[string]struct{}),
		disabledPlugins:         make(map[string]struct{}),
		forceDisabledComponents: make(map[string]struct{}),
		coreComponentsMap:       make(map[string]*CoreComponent),
		coreComponents:          make([]*CoreComponent, 0),
		pluginsMap:              make(map[string]*Plugin),
		plugins:                 make([]*Plugin, 0),
		container:               dig.New(dig.DeferAcyclicVerification()),
		log:                     nil,
		appFlagSet:              nil,
		appConfig:               nil,
		configs:                 nil,
		maskedKeys:              make([]string, 0),
		options:                 appOpts,
	}

	// provide the app itself in the container
	if err := a.container.Provide(func() *App {
		return a
	}); err != nil {
		panic(err)
	}

	// provide the app info in the container
	if err := a.container.Provide(func() *AppInfo {
		return a.appInfo
	}); err != nil {
		panic(err)
	}

	// initialize the core components and plugins
	a.init()

	return a
}

// init stage collects all parameters and loads the config files
func (a *App) init() {

	version := flag.BoolP("version", "v", false, "prints the app version")
	help := flag.BoolP("help", "h", false, "prints the app help (--full for all parameters)")
	helpFull := flag.Bool("full", false, "prints full app help (only in combination with -h)")

	if a.options.initComponent == nil {
		panic("you must configure the app with an InitComponent")
	}

	// default config
	defaultConfig := NewConfigurationSet("app", "config", "appConfigFilePath", DefaultFlagSetName, true, true, true, "config.json", "c")
	a.appFlagSet = defaultConfig.flagSet
	a.appConfig = defaultConfig.config

	a.appFlagSet.Bool(CfgAppCheckForUpdates, true, "whether to check for updates of the application or not")
	a.appFlagSet.StringSlice(CfgAppDisablePlugins, nil, "a list of plugins that shall be disabled")
	a.appFlagSet.StringSlice(CfgAppEnablePlugins, nil, "a list of plugins that shall be enabled")

	a.configs = ConfigurationSets{}
	a.configs = append(a.configs, defaultConfig)
	a.configs = append(a.configs, a.options.initComponent.AdditionalConfigs...)

	// config file flags (needed to change the path of the config files before loading them)
	configFilesFlagSet := configuration.NewUnsortedFlagSet("config_files", flag.ContinueOnError)

	for _, config := range a.configs {
		var cfgFilePath *string
		if config.shortHand != "" {
			cfgFilePath = configFilesFlagSet.StringP(config.filePathFlagName, config.shortHand, config.defaultConfigPath, fmt.Sprintf("file path of the %s configuration file", config.configName))
		} else {
			cfgFilePath = configFilesFlagSet.String(config.filePathFlagName, config.defaultConfigPath, fmt.Sprintf("file path of the %s configuration file", config.configName))
		}

		if config.filePathFlagProvideName != "" {
			// we need to provide the results of the config files flag sets, because the results are not contained in any configuration
			if err := a.container.Provide(func() *string {
				return cfgFilePath
			}, dig.Name(config.filePathFlagProvideName)); err != nil {
				panic(err)
			}
		}
	}

	// provide all config files in the container
	for cfgName, config := range a.configs.ConfigsMap() {
		c := config
		if err := a.container.Provide(func() *configuration.Configuration {
			return c
		}, dig.Name(cfgName)); err != nil {
			panic(err)
		}
	}

	//
	// Collect parameters
	//

	collectParameters := func(component *Component) {
		component.App = a

		if component.Params == nil {
			return
		}

		if component.Params.Params != nil {
			for namespace, pointerToStruct := range component.Params.Params {
				a.appConfig.BindParameters(namespace, pointerToStruct)
			}
		}

		if component.Params.AdditionalParams != nil {
			for cfgName, params := range component.Params.AdditionalParams {
				for namespace, pointerToStruct := range params {
					a.configs.ConfigsMap()[cfgName].BindParameters(namespace, pointerToStruct)
				}
			}
		}

		if component.Params.Masked != nil {
			a.maskedKeys = append(a.maskedKeys, component.Params.Masked...)
		}
	}

	collectParameters(a.options.initComponent.Component)

	forEachCoreComponent(a.options.coreComponents, func(coreComponent *CoreComponent) bool {
		collectParameters(coreComponent.Component)
		return true
	})

	forEachPlugin(a.options.plugins, func(plugin *Plugin) bool {
		collectParameters(plugin.Component)
		return true
	})

	//
	// Init Stage
	//
	// the init hook function could modify the startup behavior (e.g. to display tools)
	if a.options.initComponent.Init != nil {
		if err := a.options.initComponent.Init(a); err != nil {
			panic(fmt.Errorf("unable to initialize app: %w", err))
		}
	}

	flag.Usage = func() {
		if a.options.usageText == "" {
			// no usage text given, use default
			fmt.Fprintf(os.Stderr, `Usage of %s (%s %s):
			
Command line flags:
`, os.Args[0], a.Info().Name, a.Info().Version)
		} else {
			fmt.Fprintf(os.Stderr, a.options.usageText)
		}

		flag.PrintDefaults()
	}

	// parse command line flags from args
	configuration.ParseFlagSets(append(a.configs.FlagSets(), configFilesFlagSet))

	// check if version should be printed
	if *version {
		fmt.Println(a.Info().Name + " " + a.Info().Version)
		os.Exit(0)
	}

	// check if help text should be displayed
	if *help {
		if !*helpFull {
			// hides all non essential flags from the help/usage text.
			configuration.HideFlags(a.configs.FlagSets(), a.options.initComponent.NonHiddenFlags)
		}
		flag.Usage()
		os.Exit(0)
	}

	// load all config files
	if err := loadConfigurations(configFilesFlagSet, a.configs); err != nil {
		panic(err)
	}

	// enable version check
	a.options.versionCheckEnabled = a.appConfig.Bool(CfgAppCheckForUpdates)

	// initialize the global logger
	if err := logger.InitGlobalLogger(a.appConfig); err != nil {
		panic(err)
	}

	// initialize logger after init phase because components could modify it
	a.log = logger.NewLogger("App")
}

// printAppInfo prints app name and version info
func (a *App) printAppInfo() {

	versionString := a.Info().Version
	if _, err := goversion.NewSemver(a.Info().Version); err == nil {
		// version is a valid SemVer => release version
		versionString = "v" + versionString
	} else {
		// version is not a valid SemVer => maybe self-compiled
		versionString = "commit: " + versionString
	}

	fmt.Printf(">>>>> Starting %s %s <<<<<\n\n", a.Info().Name, versionString)
}

// prints the loaded configuration, but hides sensitive information.
func (a *App) printConfig() {
	a.appConfig.Print(a.maskedKeys)

	enablePlugins := a.appConfig.Strings(CfgAppEnablePlugins)
	disablePlugins := a.appConfig.Strings(CfgAppDisablePlugins)

	getList := func(a []string) string {
		sort.Strings(a)
		return "\n   - " + strings.Join(a, "\n   - ")
	}

	if len(enablePlugins) > 0 || len(disablePlugins) > 0 {
		if len(enablePlugins) > 0 {
			fmt.Printf("\nThe following plugins are enabled: %s\n", getList(enablePlugins))
		}
		if len(disablePlugins) > 0 {
			fmt.Printf("\nThe following plugins are disabled: %s\n", getList(disablePlugins))
		}
		fmt.Println()
	}
}

// initConfig stage
func (a *App) initConfig() {

	if a.options.initComponent.InitConfigPars != nil {
		if err := a.options.initComponent.InitConfigPars(a.container); err != nil {
			a.LogPanicf("failed to initialize init component config parameters: %s", err)
		}
	}

	forEachCoreComponent(a.options.coreComponents, func(coreComponent *CoreComponent) bool {
		if coreComponent.InitConfigPars != nil {
			if err := coreComponent.InitConfigPars(a.container); err != nil {
				a.LogPanicf("failed to initialize core component (%s) config parameters: %s", coreComponent.Name, err)
			}
		}
		return true
	})

	forEachPlugin(a.options.plugins, func(plugin *Plugin) bool {
		if plugin.InitConfigPars != nil {
			if err := plugin.InitConfigPars(a.container); err != nil {
				a.LogPanicf("failed to initialize plugin (%s) config parameters: %s", plugin.Name, err)
			}
		}
		return true
	})
}

// preProvide stage
func (a *App) preProvide() {

	initCfg := &InitConfig{
		EnabledPlugins:  a.appConfig.Strings(CfgAppEnablePlugins),
		DisabledPlugins: a.appConfig.Strings(CfgAppDisablePlugins),
	}

	if a.options.initComponent.PreProvide != nil {
		if err := a.options.initComponent.PreProvide(a.container, a, initCfg); err != nil {
			a.LogPanicf("pre-provide init component failed: %s", err)
		}
	}

	forEachCoreComponent(a.options.coreComponents, func(coreComponent *CoreComponent) bool {
		if coreComponent.PreProvide != nil {
			if err := coreComponent.PreProvide(a.container, a, initCfg); err != nil {
				a.LogPanicf("pre-provide core component (%s) failed: %s", coreComponent.Name, err)
			}
		}
		return true
	})

	forEachPlugin(a.options.plugins, func(plugin *Plugin) bool {
		if plugin.PreProvide != nil {
			if err := plugin.PreProvide(a.container, a, initCfg); err != nil {
				a.LogPanicf("pre-provide plugin (%s) failed: %s", plugin.Name, err)
			}
		}
		return true
	})

	// Enable / (Force-) disable Components
	for _, name := range initCfg.EnabledPlugins {
		a.enabledPlugins[strings.ToLower(name)] = struct{}{}
	}

	for _, name := range initCfg.DisabledPlugins {
		a.disabledPlugins[strings.ToLower(name)] = struct{}{}
	}

	for _, name := range initCfg.forceDisabledComponents {
		a.forceDisabledComponents[strings.ToLower(name)] = struct{}{}
	}
}

// addComponents stage
func (a *App) addComponents() {

	forEachCoreComponent(a.options.coreComponents, func(coreComponent *CoreComponent) bool {
		if a.isComponentForceDisabled(coreComponent.Identifier()) {
			return true
		}

		a.addCoreComponent(coreComponent)
		return true
	})

	forEachPlugin(a.options.plugins, func(plugin *Plugin) bool {
		if a.IsPluginSkipped(plugin) {
			return true
		}

		a.addPlugin(plugin)
		return true
	})

}

// provide stage
func (a *App) provide() {

	if a.options.initComponent.Provide != nil {
		if err := a.options.initComponent.Provide(a.container); err != nil {
			a.LogPanicf("provide init component failed: %s", err)
		}
	}

	a.ForEachCoreComponent(func(coreComponent *CoreComponent) bool {
		if coreComponent.Provide != nil {
			if err := coreComponent.Provide(a.container); err != nil {
				a.LogPanicf("provide core component (%s) failed: %s", coreComponent.Name, err)
			}
		}
		return true
	})

	a.ForEachPlugin(func(plugin *Plugin) bool {
		if plugin.Provide != nil {
			if err := plugin.Provide(a.container); err != nil {
				a.LogPanicf("provide plugin (%s) failed: %s", plugin.Name, err)
			}
		}
		return true
	})
}

// invoke stage
func (a *App) invoke() {

	if a.options.initComponent.DepsFunc != nil {
		if err := a.container.Invoke(a.options.initComponent.DepsFunc); err != nil {
			a.LogPanicf("invoke init component failed: %s", err)
		}
	}

	a.ForEachCoreComponent(func(coreComponent *CoreComponent) bool {
		if coreComponent.DepsFunc != nil {
			if err := a.container.Invoke(coreComponent.DepsFunc); err != nil {
				a.LogPanicf("invoke core component (%s) failed: %s", coreComponent.Name, err)
			}
		}
		return true
	})

	a.ForEachPlugin(func(plugin *Plugin) bool {
		if plugin.DepsFunc != nil {
			if err := a.container.Invoke(plugin.DepsFunc); err != nil {
				a.LogPanicf("invoke plugin (%s) failed: %s", plugin.Name, err)
			}
		}
		return true
	})
}

// configure stage
func (a *App) configure() {

	a.LogInfo("Loading core components ...")

	if a.options.initComponent.Configure != nil {
		if err := a.options.initComponent.Configure(); err != nil {
			a.LogPanicf("configure init component failed: %s", err)
		}
	}

	a.ForEachCoreComponent(func(coreComponent *CoreComponent) bool {
		if coreComponent.Configure != nil {
			if err := coreComponent.Configure(); err != nil {
				a.LogPanicf("configure core component (%s) failed: %s", coreComponent.Name, err)
			}
		}
		a.LogInfof("Loading core components: %s ... done", coreComponent.Name)
		return true
	})

	a.LogInfo("Loading plugins ...")

	a.ForEachPlugin(func(plugin *Plugin) bool {
		if plugin.Configure != nil {
			if err := plugin.Configure(); err != nil {
				a.LogPanicf("configure plugin (%s) failed: %s", plugin.Name, err)
			}
		}
		a.LogInfof("Loading plugin: %s ... done", plugin.Name)
		return true
	})
}

// initializeVersionCheck stage
func (a *App) initializeVersionCheck() {

	// do not check for updates if it was disabled
	if !a.options.versionCheckEnabled {
		return
	}

	// do not check for updates if no owner or repository was given
	if len(a.options.versionCheckOwner) == 0 || len(a.options.versionCheckRepository) == 0 {
		return
	}

	checker := version.NewVersionChecker(a.options.versionCheckOwner, a.options.versionCheckRepository, a.appInfo.Version)

	checkLatestVersion := func() {
		res, err := checker.CheckForUpdates()
		if err != nil {
			a.LogWarnf("Update check failed: %s", err)
			return
		}

		if res.Outdated {
			a.LogInfof("Update to %s available on https://github.com/%s/%s/releases/latest", a.options.versionCheckOwner, a.options.versionCheckRepository, res.Current)
			a.appInfo.LatestGitHubVersion = res.Current
		}
	}

	// execute after init
	checkLatestVersion()

	// create a background worker that checks for latest version every hour
	if err := a.Daemon().BackgroundWorker("Version update checker", func(ctx context.Context) {
		ticker := timeutil.NewTicker(checkLatestVersion, 1*time.Hour, ctx)
		ticker.WaitForGracefulShutdown()
	}, math.MaxInt16); err != nil {
		a.LogPanicf("failed to start worker: %s", err)
	}
}

// run stage
func (a *App) run() {
	a.LogInfo("Executing core components ...")

	if a.options.initComponent.Run != nil {
		if err := a.options.initComponent.Run(); err != nil {
			a.LogPanicf("run init component failed: %s", err)
		}
	}

	a.ForEachCoreComponent(func(coreComponent *CoreComponent) bool {
		if coreComponent.Run != nil {
			if err := coreComponent.Run(); err != nil {
				a.LogPanicf("run core component (%s) failed: %s", coreComponent.Name, err)
			}
		}
		a.LogInfof("Starting core component: %s ... done", coreComponent.Name)
		return true
	})

	a.LogInfo("Executing plugins ...")

	a.ForEachPlugin(func(plugin *Plugin) bool {
		if plugin.Run != nil {
			if err := plugin.Run(); err != nil {
				a.LogPanicf("run plugin (%s) failed: %s", plugin.Name, err)
			}
		}
		a.LogInfof("Starting plugin: %s ... done", plugin.Name)
		return true
	})
}

func (a *App) initializeApp() {
	a.printAppInfo()
	a.printConfig()
	a.initConfig()
	a.preProvide()
	a.addComponents()
	a.provide()
	a.invoke()
	a.configure()
	a.initializeVersionCheck()
	a.run()
}

func (a *App) Start() {
	a.initializeApp()

	a.LogInfo("Starting background workers ...")
	a.Daemon().Start()
}

func (a *App) Run() {
	a.initializeApp()

	a.LogInfo("Starting background workers ...")
	a.Daemon().Run()

	a.LogInfo("Shutdown complete!")
}

func (a *App) Shutdown() {
	a.Daemon().ShutdownAndWait()
}

func (a *App) Info() *AppInfo {
	return a.appInfo
}

func (a *App) Config() *configuration.Configuration {
	return a.appConfig
}

func (a *App) FlagSet() *flag.FlagSet {
	return a.appFlagSet
}

func (a *App) AdditionalConfigs() map[string]*configuration.Configuration {
	return a.configs.ConfigsMap()
}

func (a *App) AdditionalFlagSets() map[string]*flag.FlagSet {
	return a.configs.FlagSetsMap()
}

func (a *App) Daemon() daemon.Daemon {
	return a.options.daemon
}

func (a *App) addCoreComponent(coreComponent *CoreComponent) {
	name := coreComponent.Name

	if _, exists := a.coreComponentsMap[name]; exists {
		panic("duplicate core component - \"" + name + "\" was defined already")
	}

	a.coreComponentsMap[name] = coreComponent
	a.coreComponents = append(a.coreComponents, coreComponent)
}

func (a *App) addPlugin(plugin *Plugin) {
	name := plugin.Name

	if _, exists := a.pluginsMap[name]; exists {
		panic("duplicate plugin - \"" + name + "\" was defined already")
	}

	a.pluginsMap[name] = plugin
	a.plugins = append(a.plugins, plugin)
}

func (a *App) isPluginEnabled(identifier string) bool {
	_, exists := a.enabledPlugins[identifier]
	return exists
}

func (a *App) isPluginDisabled(identifier string) bool {
	_, exists := a.disabledPlugins[identifier]
	return exists
}

func (a *App) isComponentForceDisabled(identifier string) bool {
	_, exists := a.forceDisabledComponents[identifier]
	return exists
}

// IsPluginSkipped returns whether the plugin is loaded or skipped.
func (a *App) IsPluginSkipped(plugin *Plugin) bool {
	// list of disabled plugins has the highest priority
	if a.isPluginDisabled(plugin.Identifier()) || a.isComponentForceDisabled(plugin.Identifier()) {
		return true
	}

	// if the plugin was not in the list of disabled plugins, it is only skipped if
	// the plugin was not enabled and not in the list of enabled plugins.
	return plugin.Status != StatusEnabled && !a.isPluginEnabled(plugin.Identifier())
}

// CoreComponentForEachFunc is used in ForEachCoreComponent.
// Returning false indicates to stop looping.
type CoreComponentForEachFunc func(coreComponent *CoreComponent) bool

func forEachCoreComponent(coreComponents []*CoreComponent, f CoreComponentForEachFunc) {
	for _, coreComponent := range coreComponents {
		if !f(coreComponent) {
			break
		}
	}
}

// ForEachCoreComponent calls the given CoreComponentForEachFunc on each loaded core components.
func (a *App) ForEachCoreComponent(f CoreComponentForEachFunc) {
	forEachCoreComponent(a.coreComponents, f)
}

// PluginForEachFunc is used in ForEachPlugin.
// Returning false indicates to stop looping.
type PluginForEachFunc func(plugin *Plugin) bool

func forEachPlugin(plugins []*Plugin, f PluginForEachFunc) {
	for _, plugin := range plugins {
		if !f(plugin) {
			break
		}
	}
}

// ForEachPlugin calls the given PluginForEachFunc on each loaded plugin.
func (n *App) ForEachPlugin(f PluginForEachFunc) {
	forEachPlugin(n.plugins, f)
}

//
// Logger
//

// LogDebug uses fmt.Sprint to construct and log a message.
func (a *App) LogDebug(args ...interface{}) {
	a.log.Debug(args...)
}

// LogDebugf uses fmt.Sprintf to log a templated message.
func (a *App) LogDebugf(template string, args ...interface{}) {
	a.log.Debugf(template, args...)
}

// LogError uses fmt.Sprint to construct and log a message.
func (a *App) LogError(args ...interface{}) {
	a.log.Error(args...)
}

// LogErrorf uses fmt.Sprintf to log a templated message.
func (a *App) LogErrorf(template string, args ...interface{}) {
	a.log.Errorf(template, args...)
}

// LogFatal uses fmt.Sprint to construct and log a message, then calls os.Exit.
func (a *App) LogFatal(args ...interface{}) {
	a.log.Fatal(args...)
}

// LogFatalf uses fmt.Sprintf to log a templated message, then calls os.Exit.
func (a *App) LogFatalf(template string, args ...interface{}) {
	a.log.Fatalf(template, args...)
}

// LogInfo uses fmt.Sprint to construct and log a message.
func (a *App) LogInfo(args ...interface{}) {
	a.log.Info(args...)
}

// LogInfof uses fmt.Sprintf to log a templated message.
func (a *App) LogInfof(template string, args ...interface{}) {
	a.log.Infof(template, args...)
}

// LogWarn uses fmt.Sprint to construct and log a message.
func (a *App) LogWarn(args ...interface{}) {
	a.log.Warn(args...)
}

// LogWarnf uses fmt.Sprintf to log a templated message.
func (a *App) LogWarnf(template string, args ...interface{}) {
	a.log.Warnf(template, args...)
}

// LogPanic uses fmt.Sprint to construct and log a message, then panics.
func (a *App) LogPanic(args ...interface{}) {
	a.log.Panic(args...)
}

// LogPanicf uses fmt.Sprintf to log a templated message, then panics.
func (a *App) LogPanicf(template string, args ...interface{}) {
	a.log.Panicf(template, args...)
}
