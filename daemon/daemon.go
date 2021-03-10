package daemon

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"sync"

	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/hive.go/syncutils"
	"github.com/iotaledger/hive.go/typeutils"
)

// Errors for the daemon package
var (
	ErrDaemonAlreadyStopped                 = errors.New("daemon was already stopped")
	ErrExistingBackgroundWorkerStillRunning = errors.New("existing background worker is still running")
)

// functions kept for backwards compatibility
var defaultDaemon = New()

// GetRunningBackgroundWorkers gets the running background workers of the default daemon instance.
func GetRunningBackgroundWorkers() []string {
	return defaultDaemon.GetRunningBackgroundWorkers()
}

// BackgroundWorker adds a new background worker to the default daemon instance. Use shutdownOrderWorker
// to define in which shutdown order this particular background worker is shut down (higher = earlier).
func BackgroundWorker(name string, handler WorkerFunc, priority ...int) error {
	return defaultDaemon.BackgroundWorker(name, handler, priority...)
}

// DebugEnabled allows to configure the daemon to issue log messages for debugging purposes.
func DebugEnabled(enabled bool) {
	defaultDaemon.DebugEnabled(enabled)
}

// Start starts the default daemon instance.
func Start() {
	defaultDaemon.Start()
}

// Run runs the default daemon instance and then waits for the daemon to shutdown.
func Run() {
	defaultDaemon.Run()
}

// Shutdown signals all background worker of the default daemon instance to shut down.
// This call doesn't await termination of the background workers.
func Shutdown() {
	defaultDaemon.Shutdown()
}

// ShutdownAndWait signals all background worker of the default daemon instance to shut down and
// then waits for their termination.
func ShutdownAndWait() {
	defaultDaemon.ShutdownAndWait()
}

// IsRunning checks whether the default daemon instance is running.
func IsRunning() bool {
	return defaultDaemon.IsRunning()
}

// IsStopped checks whether the default daemon instance was stopped.
func IsStopped() bool {
	return defaultDaemon.IsStopped()
}

// ContextStopped returns a context that is done when the deamon is stopped.
func ContextStopped() context.Context {
	return defaultDaemon.ContextStopped()
}

// New creates a new daemon instance.
func New() *OrderedDaemon {
	stoppedCtx, stoppedCtxCancel := context.WithCancel(context.Background())

	return &OrderedDaemon{
		running:                typeutils.NewAtomicBool(),
		stopped:                typeutils.NewAtomicBool(),
		stoppedCtx:             stoppedCtx,
		stoppedCtxCancel:       stoppedCtxCancel,
		workers:                make(map[string]*worker),
		shutdownOrderWorker:    make([]string, 0),
		wgPerSameShutdownOrder: make(map[int]*sync.WaitGroup),
	}
}

// OrderedDaemon is an orchestrator for background workers.
// stopOnce ensures that the daemon can only be terminated once.
type OrderedDaemon struct {
	running                *typeutils.AtomicBool
	stopped                *typeutils.AtomicBool
	stoppedCtx             context.Context
	stoppedCtxCancel       context.CancelFunc
	stopOnce               sync.Once
	workers                map[string]*worker
	shutdownOrderWorker    []string
	wgPerSameShutdownOrder map[int]*sync.WaitGroup
	lock                   syncutils.RWMutex
	logger                 *logger.Logger
}

type worker struct {
	handler        WorkerFunc
	running        *typeutils.AtomicBool
	shutdownOrder  int
	shutdownSignal chan struct{}
}

// GetRunningBackgroundWorkers gets the running background workers sorted by their priority.
func (d *OrderedDaemon) GetRunningBackgroundWorkers() []string {
	d.lock.RLock()
	defer d.lock.RUnlock()

	result := make([]string, 0)
	for _, name := range d.shutdownOrderWorker {
		if !d.workers[name].running.IsSet() {
			continue
		}
		result = append(result, name)
	}

	for i, j := 0, len(result)-1; i < j; i, j = i+1, j-1 {
		result[i], result[j] = result[j], result[i]
	}

	return result
}

func (d *OrderedDaemon) runBackgroundWorker(name string, backgroundWorker WorkerFunc) {
	worker := d.workers[name]
	shutdownOrderWaitGroup := d.wgPerSameShutdownOrder[worker.shutdownOrder]
	shutdownOrderWaitGroup.Add(1)

	worker.running.Set()
	go func() {
		if d.logger != nil {
			d.logger.Debugf("Starting Background Worker: %s ...", name)
		}
		backgroundWorker(worker.shutdownSignal)
		worker.running.UnSet()
		shutdownOrderWaitGroup.Done()
		if d.logger != nil {
			d.logger.Debugf("Stopping Background Worker: %s ... done", name)
		}
	}()
}

// BackgroundWorker adds a new background worker to the daemon.
// Use order to define in which shutdown order this particular
// background worker is shut down (higher = earlier).
func (d *OrderedDaemon) BackgroundWorker(name string, handler WorkerFunc, order ...int) error {

	if d.IsStopped() {
		return ErrDaemonAlreadyStopped
	}

	d.lock.Lock()
	defer d.lock.Unlock()

	exWorker, has := d.workers[name]
	if has {
		if exWorker.running.IsSet() {
			return fmt.Errorf("%w: %s is still running", ErrExistingBackgroundWorkerStillRunning, name)
		}

		// remove the existing worker from the shutdown order
		for i, exName := range d.shutdownOrderWorker {
			if exName != name {
				continue
			}
			if i < len(d.shutdownOrderWorker)-1 {
				copy(d.shutdownOrderWorker[i:], d.shutdownOrderWorker[i+1:])
			}
			d.shutdownOrderWorker[len(d.shutdownOrderWorker)-1] = ""
			d.shutdownOrderWorker = d.shutdownOrderWorker[:len(d.shutdownOrderWorker)-1]
		}
	}

	var shutdownOrder int
	if len(order) > 0 && order[0] != 0 {
		shutdownOrder = order[0]
	} else {
		shutdownOrder = 0
	}

	if _, ok := d.wgPerSameShutdownOrder[shutdownOrder]; !ok {
		d.wgPerSameShutdownOrder[shutdownOrder] = &sync.WaitGroup{}
	}

	d.workers[name] = &worker{
		handler:        handler,
		running:        typeutils.NewAtomicBool(),
		shutdownOrder:  shutdownOrder,
		shutdownSignal: make(chan struct{}, 1),
	}

	// add to the shutdown sequence and order by order
	d.shutdownOrderWorker = append(d.shutdownOrderWorker, name)

	// must be done while holding the lock
	sort.Slice(d.shutdownOrderWorker, func(i, j int) bool {
		return d.workers[d.shutdownOrderWorker[i]].shutdownOrder > d.workers[d.shutdownOrderWorker[j]].shutdownOrder
	})

	if d.IsRunning() {
		d.runBackgroundWorker(name, handler)
	}

	return nil
}

// DebugEnabled allows to configure the daemon to issue log messages for debugging purposes.
func (d *OrderedDaemon) DebugEnabled(enabled bool) {
	if enabled {
		defaultDaemon.logger = logger.NewLogger("Daemon")
	} else {
		defaultDaemon.logger = nil
	}
}

// Start starts the daemon.
func (d *OrderedDaemon) Start() {
	// do not allow restarts
	if d.IsStopped() {
		return
	}

	d.lock.Lock()
	defer d.lock.Unlock()

	if !d.IsRunning() {
		d.running.Set()
		for name, worker := range d.workers {
			d.runBackgroundWorker(name, worker.handler)
		}
	}
}

// Run runs the daemon and then waits for the daemon to shutdown.
func (d *OrderedDaemon) Run() {
	d.Start()
	if wg := d.waitGroupForLastPriority(); wg != nil {
		wg.Wait()
	}
}

// returns on the waitGroup of the lowest shutdownOrderWorker or nil if not workers.
func (d *OrderedDaemon) waitGroupForLastPriority() *sync.WaitGroup {
	d.lock.Lock()
	defer d.lock.Unlock()

	if len(d.shutdownOrderWorker) == 0 {
		return nil
	}
	// find lowest shutdown priority
	lowestShutdownPriorityWorker := d.workers[d.shutdownOrderWorker[len(d.shutdownOrderWorker)-1]]
	// return waitGroup for lowest priority
	return d.wgPerSameShutdownOrder[lowestShutdownPriorityWorker.shutdownOrder]
}

func (d *OrderedDaemon) shutdown() {
	if d.logger != nil {
		d.logger.Debugf("Shutting down ...")
	}

	d.stopped.Set()
	d.stoppedCtxCancel()
	if !d.IsRunning() {
		return
	}

	d.stopWorkers()
	d.running.UnSet()
	d.clear()
}

// stopWorkers stops all the workers of the daemon
func (d *OrderedDaemon) stopWorkers() {
	d.lock.RLock()
	defer d.lock.RUnlock()

	// stop all the workers
	if len(d.shutdownOrderWorker) > 0 {
		// initialize with the priority of the first worker
		prevPriority := d.workers[d.shutdownOrderWorker[0]].shutdownOrder
		for _, name := range d.shutdownOrderWorker {
			worker := d.workers[name]
			if !worker.running.IsSet() {
				// the worker's shutdown channel will be automatically garbage collected
				continue
			}
			// if the current worker has a lower priority...
			if worker.shutdownOrder < prevPriority {
				// wait for every worker in the previous shutdown priority to terminate
				d.wgPerSameShutdownOrder[prevPriority].Wait()
				prevPriority = worker.shutdownOrder
			}
			if d.logger != nil {
				d.logger.Debugf("Stopping Background Worker: %s ...", name)
			}
			close(worker.shutdownSignal)
		}
		// wait for the last priority to finish
		d.wgPerSameShutdownOrder[prevPriority].Wait()
	}
}

// clear clears the daemon
func (d *OrderedDaemon) clear() {
	d.lock.Lock()
	defer d.lock.Unlock()

	d.workers = nil
	d.shutdownOrderWorker = nil
	d.wgPerSameShutdownOrder = nil
}

// Shutdown signals all background worker of the daemon shut down.
// This call doesn't await termination of the background workers.
func (d *OrderedDaemon) Shutdown() {
	go d.stopOnce.Do(d.shutdown)
}

// ShutdownAndWait signals all background worker of the daemon to shut down and then waits for their termination.
func (d *OrderedDaemon) ShutdownAndWait() {
	d.stopOnce.Do(d.shutdown)
}

// IsRunning checks whether the daemon is running.
func (d *OrderedDaemon) IsRunning() bool {
	return d.running.IsSet()
}

// IsStopped checks whether the daemon was stopped.
func (d *OrderedDaemon) IsStopped() bool {
	return d.stopped.IsSet()
}

// ContextStopped returns a context that is done when the deamon is stopped.
func (d *OrderedDaemon) ContextStopped() context.Context {
	return d.stoppedCtx
}
