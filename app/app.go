package app

import (
	"context"
	"fmt"
	"math"
	"os"
	"runtime/debug"
	"sort"
	"strings"
	"time"

	goversion "github.com/hashicorp/go-version"
	flag "github.com/spf13/pflag"
	"go.uber.org/dig"

	"github.com/iotaledger/hive.go/app/configuration"
	"github.com/iotaledger/hive.go/app/daemon"
	"github.com/iotaledger/hive.go/app/version"
	"github.com/iotaledger/hive.go/ierrors"
	"github.com/iotaledger/hive.go/lo"
	"github.com/iotaledger/hive.go/log"
	"github.com/iotaledger/hive.go/runtime/timeutil"
	"github.com/iotaledger/hive.go/runtime/workerpool"
)

const (
	DefaultFlagSetName = "appConfig"
)

// Info provides information about the app.
type Info struct {
	Name                string
	Version             string
	LatestGitHubVersion string
}

type ParametersApp struct {
	CheckForUpdates bool `default:"true" usage:"whether to check for updates of the application or not"`
}

type App struct {
	log.Logger

	appInfo                *Info
	componentsEnabledState map[string]bool
	componentsMap          map[string]*Component
	components             []*Component
	container              *dig.Container
	loggerRoot             log.Logger
	appFlagSet             *flag.FlagSet
	appConfig              *configuration.Configuration
	appParams              *ParametersApp
	configs                ConfigurationSets
	maskedKeys             []string
	options                *Options
}

func New(name string, version string, optionalOptions ...Option) *App {
	appOpts := &Options{}
	appOpts.apply(defaultOptions...)
	appOpts.apply(optionalOptions...)

	if strings.HasPrefix(strings.ToLower(version), "v") {
		if _, err := goversion.NewSemver(version[1:]); err == nil {
			// version is a valid SemVer with a "v" prefix => remove the "v" prefix
			version = version[1:]
		}
	}

	if version == "" {
		panic("unable to initialize app: no version given")
	}

	a := &App{
		appInfo: &Info{
			Name:                name,
			Version:             version,
			LatestGitHubVersion: "",
		},
		componentsEnabledState: make(map[string]bool),
		componentsMap:          make(map[string]*Component),
		components:             make([]*Component, 0),
		container:              dig.New(dig.DeferAcyclicVerification()),
		loggerRoot:             nil,
		appFlagSet:             nil,
		appConfig:              nil,
		configs:                nil,
		maskedKeys:             make([]string, 0),
		options:                appOpts,
	}

	// provide the app itself in the container
	if err := a.container.Provide(func() *App {
		return a
	}); err != nil {
		panic(err)
	}

	// provide the app info in the container
	if err := a.container.Provide(func() *Info {
		return a.appInfo
	}); err != nil {
		panic(err)
	}

	// initialize the components
	a.init()

	return a
}

// init stage collects all parameters and loads the config files.
func (a *App) init() {
	version := flag.BoolP("version", "v", false, "prints the app version")
	help := flag.BoolP("help", "h", false, "prints the app help (--full for all parameters)")
	helpFull := flag.Bool("full", false, "prints full app help (only in combination with -h)")

	if a.options.initComponent == nil {
		panic("unable to initialize app: no InitComponent given")
	}

	// default config
	defaultConfig := NewConfigurationSet("app", "config", "appConfigFilePath", DefaultFlagSetName, true, true, true, "config.json", "c")
	a.appFlagSet = defaultConfig.flagSet
	a.appConfig = defaultConfig.config

	a.appParams = &ParametersApp{}
	a.appConfig.BindParameters(a.appFlagSet, "app", a.appParams)

	loggerConfig := &LoggerConfig{}
	a.appConfig.BindParameters(a.appFlagSet, "logger", loggerConfig)

	// provide the app params in the container
	if err := a.container.Provide(func() *ParametersApp {
		return a.appParams
	}); err != nil {
		panic(err)
	}

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
		if err := a.container.Provide(func() *configuration.Configuration {
			return config
		}, dig.Name(cfgName)); err != nil {
			panic(err)
		}
	}

	//
	// Collect parameters
	//

	collectParameters := func(component *Component) {
		component.app = a

		if component.Params == nil {
			return
		}

		if component.Params.Params != nil {
			// sort namespaces first
			sortedNamespaces := make([]string, 0, len(component.Params.Params))
			for namespace := range component.Params.Params {
				sortedNamespaces = append(sortedNamespaces, namespace)
			}

			sort.Slice(sortedNamespaces, func(i, j int) bool {
				return sortedNamespaces[i] < sortedNamespaces[j]
			})

			// bind parameters in sorted order
			for _, namespace := range sortedNamespaces {
				pointerToStruct := component.Params.Params[namespace]
				a.appConfig.BindParameters(a.appFlagSet, namespace, pointerToStruct)
			}
		}

		if component.Params.AdditionalParams != nil {
			// sort config names first
			sortedCfgNames := make([]string, 0, len(component.Params.AdditionalParams))
			for cfgName := range component.Params.AdditionalParams {
				sortedCfgNames = append(sortedCfgNames, cfgName)
			}

			sort.Slice(sortedCfgNames, func(i, j int) bool {
				return sortedCfgNames[i] < sortedCfgNames[j]
			})

			// iterate through config names in sorted order
			for _, cfgName := range sortedCfgNames {
				params := component.Params.AdditionalParams[cfgName]

				// sort namespaces first
				sortedNamespaces := make([]string, 0, len(params))
				for namespace := range params {
					sortedNamespaces = append(sortedNamespaces, namespace)
				}

				sort.Slice(sortedNamespaces, func(i, j int) bool {
					return sortedNamespaces[i] < sortedNamespaces[j]
				})

				// bind parameters in sorted order
				for _, namespace := range sortedNamespaces {
					pointerToStruct := params[namespace]
					a.configs.ConfigsMap()[cfgName].BindParameters(a.configs.FlagSetsMap()[cfgName], namespace, pointerToStruct)
				}
			}
		}

		if component.Params.Masked != nil {
			a.maskedKeys = append(a.maskedKeys, component.Params.Masked...)
		}
	}

	collectParameters(a.options.initComponent.Component)

	forEachComponent(a.options.components, func(component *Component) bool {
		collectParameters(component)

		return true
	})

	//
	// Init Stage
	//
	// the init hook function could modify the startup behavior (e.g. to display tools)
	if a.options.initComponent.Init != nil {
		if err := a.options.initComponent.Init(a); err != nil {
			panic(ierrors.Wrap(err, "unable to initialize app"))
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
			// hides all non-essential flags from the help/usage text.
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
	a.options.versionCheckEnabled = a.appParams.CheckForUpdates

	// initialize the root logger
	loggerRoot, err := NewLoggerFromConfig(loggerConfig)
	if err != nil {
		panic(err)
	}
	a.loggerRoot = loggerRoot

	// initialize logger after init phase because components could modify it
	a.Logger = a.loggerRoot.NewChildLogger("App")

	// initialize the loggers of the components
	forEachComponent(a.options.components, func(component *Component) bool {
		component.Logger = a.loggerRoot.NewChildLogger(component.Name)

		return true
	})
}

// printAppInfo prints app name and version info.
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

	componentsByID := lo.KeyBy(a.options.components, func(c *Component) string {
		return c.Identifier()
	})

	enabledComponents := []string{}
	disabledComponents := []string{}
	for componentID, enabled := range a.componentsEnabledState {
		component, exists := componentsByID[componentID]
		if !exists {
			continue
		}

		if enabled {
			enabledComponents = append(enabledComponents, component.Name)
		} else {
			disabledComponents = append(disabledComponents, component.Name)
		}
	}

	getList := func(a []string) string {
		sort.Strings(a)

		return "\n   - " + strings.Join(a, "\n   - ")
	}

	if len(enabledComponents) > 0 || len(disabledComponents) > 0 {
		if len(enabledComponents) > 0 {
			fmt.Printf("\nThe following components are enabled: %s\n", getList(enabledComponents))
		}
		if len(disabledComponents) > 0 {
			fmt.Printf("\nThe following components are disabled: %s\n", getList(disabledComponents))
		}
		fmt.Println()
	}
}

// initConfig stage.
func (a *App) initConfig() {
	if a.options.initComponent.InitConfigParams != nil {
		if err := a.options.initComponent.InitConfigParams(a.container); err != nil {
			a.LogPanicf("failed to initialize init component config parameters: %s", err)
		}
	}

	forEachComponent(a.options.components, func(component *Component) bool {
		if component.InitConfigParams != nil {
			if err := component.InitConfigParams(a.container); err != nil {
				a.LogPanicf("failed to initialize component (%s) config parameters: %s", component.Name, err)
			}
		}

		return true
	})
}

// preProvide stage.
func (a *App) preProvide() {
	forEachComponent(a.options.components, func(component *Component) bool {
		// Enable / disable Components
		// If no "IsEnabled" function is given, components are enabled by default.
		a.componentsEnabledState[component.Identifier()] = component.IsEnabled == nil || component.IsEnabled(a.container)

		return true
	})
}

// addComponents stage.
func (a *App) addComponents() {
	forEachComponent(a.options.components, func(component *Component) bool {
		if !a.IsComponentEnabled(component.Identifier()) {
			return true
		}

		component.WorkerPool = workerpool.New(fmt.Sprintf("Component-%s", component.Name), workerpool.WithWorkerCount(1))
		a.addComponent(component)

		return true
	})
}

// provide stage.
func (a *App) provide() {
	if a.options.initComponent.Provide != nil {
		if err := a.options.initComponent.Provide(a.container); err != nil {
			a.LogPanicf("provide init component failed: %s", err)
		}
	}

	a.ForEachComponent(func(component *Component) bool {
		if component.Provide != nil {
			if err := component.Provide(a.container); err != nil {
				a.LogPanicf("provide component (%s) failed: %s", component.Name, err)
			}
		}

		return true
	})
}

// invoke stage.
func (a *App) invoke() {
	if a.options.initComponent.DepsFunc != nil {
		if err := a.container.Invoke(a.options.initComponent.DepsFunc); err != nil {
			a.LogPanicf("invoke init component failed: %s", err)
		}
	}

	a.ForEachComponent(func(component *Component) bool {
		if component.DepsFunc != nil {
			if err := a.container.Invoke(component.DepsFunc); err != nil {
				a.LogPanicf("invoke component (%s) failed: %s", component.Name, err)
			}
		}

		return true
	})
}

// configure stage.
func (a *App) configure() {
	a.LogInfo("Loading components ...")

	if a.options.initComponent.Configure != nil {
		if err := a.options.initComponent.Configure(); err != nil {
			a.LogPanicf("configure init component failed: %s", err)
		}
	}

	a.ForEachComponent(func(component *Component) bool {
		if component.Configure != nil {
			if err := component.Configure(); err != nil {
				a.LogPanicf("configure component (%s) failed: %s", component.Name, err)
			}
		}
		a.LogInfof("Loading components: %s ... done", component.Name)

		return true
	})
}

// initializeVersionCheck stage.
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
			a.LogInfof("Update to %s %s available on https://github.com/%s/%s/releases/latest",
				a.options.versionCheckRepository, res.Current, a.options.versionCheckOwner, a.options.versionCheckRepository)
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

// run stage.
func (a *App) run() {
	a.LogInfo("Executing components ...")

	if a.options.initComponent.Run != nil {
		if err := a.options.initComponent.Run(); err != nil {
			a.LogPanicf("run init component failed: %s", err)
		}
	}

	a.ForEachComponent(func(component *Component) bool {
		a.LogInfof("Starting component: %s ...", component.Name)
		component.WorkerPool.Start()
		if component.Run != nil {
			if err := component.Run(); err != nil {
				a.LogPanicf("run component (%s) failed: %s", component.Name, err)
			}
		}

		return true
	})
}

func (a *App) initializeApp() {
	a.printAppInfo()
	a.initConfig()
	a.preProvide()
	a.printConfig()
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
	defer func() {
		r := recover()
		if r != nil {
			if err, ok := r.(error); ok {
				a.LogPanicf("application panic, err: %s \n %s", err.Error(), string(debug.Stack()))
			}
			a.LogPanicf("application panic: %v \n %s", r, string(debug.Stack()))
		}
	}()
	a.initializeApp()

	a.LogInfo("Starting background workers ...")
	a.Daemon().Run()

	a.ForEachComponent(func(component *Component) bool {
		component.WorkerPool.Shutdown()

		return true
	})

	a.LogInfo("Shutdown complete!")

	a.loggerRoot.Shutdown()
}

func (a *App) Shutdown() {
	a.Daemon().ShutdownAndWait()

	a.LogInfo("Shutdown complete!")

	a.loggerRoot.Shutdown()
}

func (a *App) Info() *Info {
	return a.appInfo
}

func (a *App) Config() *configuration.Configuration {
	return a.appConfig
}

func (a *App) FlagSet() *flag.FlagSet {
	return a.appFlagSet
}

func (a *App) Parameters() *ParametersApp {
	return a.appParams
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

func (a *App) addComponent(component *Component) {
	identifier := component.Identifier()

	if _, exists := a.componentsMap[identifier]; exists {
		panic("duplicate component - \"" + component.Name + "\" was defined already")
	}

	a.componentsMap[identifier] = component
	a.components = append(a.components, component)
}

// IsComponentEnabled returns whether the component is enabled.
func (a *App) IsComponentEnabled(identifier string) bool {
	enabled, exists := a.componentsEnabledState[identifier]

	return exists && enabled
}

// ComponentForEachFunc is used in ForEachComponent.
// Returning false indicates to stop looping.
type ComponentForEachFunc func(component *Component) bool

func forEachComponent(components []*Component, f ComponentForEachFunc) {
	for _, component := range components {
		if !f(component) {
			break
		}
	}
}

// ForEachComponent calls the given ComponentForEachFunc on each loaded component.
func (a *App) ForEachComponent(f ComponentForEachFunc) {
	forEachComponent(a.components, f)
}
