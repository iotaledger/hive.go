package daemon

import (
	"context"

	"github.com/iotaledger/hive.go/log"
)

// WorkerFunc is the function to run a worker accepting its context.
type WorkerFunc = func(ctx context.Context)

// Daemon specifies an interface to run background go routines.
type Daemon interface {
	// GetRunningBackgroundWorkers gets the running background workers.
	GetRunningBackgroundWorkers() []string

	// BackgroundWorker adds a new background worker to the daemon.
	// Use order to define in which shutdown order this particular
	// background worker is shut down (higher = earlier).
	BackgroundWorker(name string, handler WorkerFunc, order ...int) error

	// DebugLogger allows to pass a logger to the daemon to issue log messages for debugging purposes.
	DebugLogger(logger log.Logger)

	// Start starts the daemon.
	Start()

	// Run runs the daemon and then waits for the daemon to shutdown.
	Run()

	// Shutdown signals all background worker of the daemon shut down.
	// This call doesn't await termination of the background workers.
	Shutdown()

	// Shutdown signals all background worker of the daemon to shut down and
	// then waits for their termination.
	ShutdownAndWait()

	// IsRunning checks whether the daemon is running.
	IsRunning() bool

	// IsStopped checks whether the daemon was stopped.
	IsStopped() bool

	// ContextStopped returns a context that is done when the deamon is stopped.
	ContextStopped() context.Context
}
