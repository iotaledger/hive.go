package workerpool

import (
	"runtime"
)

var defaultOptions = &Options{
	Alias:                "",
	WorkerCount:          2 * runtime.NumCPU(),
	QueueSize:            4 * runtime.NumCPU(),
	FlushTasksAtShutdown: false,
}

func WithAlias(alias string) Option {
	return func(o *Options) {
		o.Alias = alias
	}
}

func WorkerCount(workerCount int) Option {
	return func(args *Options) {
		args.WorkerCount = workerCount
	}
}

func QueueSize(queueSize int) Option {
	return func(args *Options) {
		args.QueueSize = queueSize
	}
}

func FlushTasksAtShutdown(flush bool) Option {
	return func(args *Options) {
		args.FlushTasksAtShutdown = flush
	}
}

type Options struct {
	Alias                string
	WorkerCount          int
	QueueSize            int
	FlushTasksAtShutdown bool
}

func (options Options) Override(optionalOptions ...Option) *Options {
	result := &options
	for _, option := range optionalOptions {
		option(result)
	}

	return result
}

type Option func(*Options)
