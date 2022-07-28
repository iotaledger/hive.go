package app

import (
	daemon2 "github.com/iotaledger/hive.go/core/daemon"
)

// the default options applied to the App.
var defaultAppOptions = []AppOption{
	WithDaemon(daemon2.New()),
}

// AppOptions defines options for an App.
type AppOptions struct {
	daemon                 daemon2.Daemon
	initComponent          *InitComponent
	coreComponents         []*CoreComponent
	plugins                []*Plugin
	versionCheckEnabled    bool
	versionCheckOwner      string
	versionCheckRepository string
	usageText              string
}

// AppOption is a function setting a AppOptions option.
type AppOption func(opts *AppOptions)

// applies the given AppOption.
func (no *AppOptions) apply(opts ...AppOption) {
	for _, opt := range opts {
		opt(no)
	}
}

// WithInitComponent sets the init component.
func WithInitComponent(initComponent *InitComponent) AppOption {
	return func(opts *AppOptions) {
		opts.initComponent = initComponent
	}
}

// WithDaemon sets the used daemon.
func WithDaemon(daemon daemon2.Daemon) AppOption {
	return func(args *AppOptions) {
		args.daemon = daemon
	}
}

// WithCoreComponents sets the core components.
func WithCoreComponents(coreComponents ...*CoreComponent) AppOption {
	return func(args *AppOptions) {
		args.coreComponents = append(args.coreComponents, coreComponents...)
	}
}

// WithPlugins sets the plugins.
func WithPlugins(plugins ...*Plugin) AppOption {
	return func(args *AppOptions) {
		args.plugins = append(args.plugins, plugins...)
	}
}

// WithVersionCheck enables the GitHub version check.
func WithVersionCheck(owner string, repository string) AppOption {
	return func(opts *AppOptions) {
		opts.versionCheckOwner = owner
		opts.versionCheckRepository = repository
	}
}

// WithUsageText overwrites the default usage text of the app.
func WithUsageText(usageText string) AppOption {
	return func(opts *AppOptions) {
		opts.usageText = usageText
	}
}
