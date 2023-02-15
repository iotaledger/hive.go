package app

import (
	"github.com/iotaledger/hive.go/app/daemon"
)

// the default options applied to the App.
var defaultOptions = []Option{
	WithDaemon(daemon.New()),
}

// Options defines options for an App.
type Options struct {
	daemon                 daemon.Daemon
	initComponent          *InitComponent
	coreComponents         []*CoreComponent
	plugins                []*Plugin
	versionCheckEnabled    bool
	versionCheckOwner      string
	versionCheckRepository string
	usageText              string
}

// Option is a function setting a Options option.
type Option func(opts *Options)

// applies the given Option.
func (no *Options) apply(opts ...Option) {
	for _, opt := range opts {
		opt(no)
	}
}

// WithInitComponent sets the init component.
func WithInitComponent(initComponent *InitComponent) Option {
	return func(opts *Options) {
		opts.initComponent = initComponent
	}
}

// WithDaemon sets the used daemon.
func WithDaemon(d daemon.Daemon) Option {
	return func(args *Options) {
		args.daemon = d
	}
}

// WithCoreComponents sets the core components.
func WithCoreComponents(coreComponents ...*CoreComponent) Option {
	return func(args *Options) {
		args.coreComponents = append(args.coreComponents, coreComponents...)
	}
}

// WithPlugins sets the plugins.
func WithPlugins(plugins ...*Plugin) Option {
	return func(args *Options) {
		args.plugins = append(args.plugins, plugins...)
	}
}

// WithVersionCheck enables the GitHub version check.
func WithVersionCheck(owner string, repository string) Option {
	return func(opts *Options) {
		opts.versionCheckOwner = owner
		opts.versionCheckRepository = repository
	}
}

// WithUsageText overwrites the default usage text of the app.
func WithUsageText(usageText string) Option {
	return func(opts *Options) {
		opts.usageText = usageText
	}
}
