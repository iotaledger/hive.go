package daemon

import (
	"context"
	"sort"
	"sync"
	"sync/atomic"

	"github.com/iotaledger/hive.go/ierrors"
	"github.com/iotaledger/hive.go/log"
	"github.com/iotaledger/hive.go/runtime/syncutils"
)

// Errors for the daemon package.
var (
	ErrDaemonAlreadyStopped                 = ierrors.New("daemon was already stopped")
	ErrDuplicateBackgroundWorker            = ierrors.New("duplicate background worker")
	ErrExistingBackgroundWorkerStillRunning = ierrors.New("existing background worker is still running")
)

// functions kept for backwards compatibility.
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

// DebugLogger allows to pass a logger to the daemon to issue log messages for debugging purposes.
func DebugLogger(logger log.Logger) {
	defaultDaemon.DebugLogger(logger)
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
	running                atomic.Bool
	stopped                atomic.Bool
	stoppedCtx             context.Context
	stoppedCtxCancel       context.CancelFunc
	stopOnce               sync.Once
	workers                map[string]*worker
	shutdownOrderWorker    []string
	wgPerSameShutdownOrder map[int]*sync.WaitGroup
	lock                   syncutils.RWMutex
	logger                 log.Logger
}

type worker struct {
	ctx           context.Context
	ctxCancel     context.CancelFunc
	handler       WorkerFunc
	running       atomic.Bool
	shutdownOrder int
}

// GetRunningBackgroundWorkers gets the running background workers sorted by their priority.
func (d *OrderedDaemon) GetRunningBackgroundWorkers() []string {
	d.lock.RLock()
	defer d.lock.RUnlock()

	result := make([]string, 0)
	for _, name := range d.shutdownOrderWorker {
		if !d.workers[name].running.Load() {
			continue
		}
		result = append(result, name)
	}

	// reverse order
	for i, j := 0, len(result)-1; i < j; i, j = i+1, j-1 {
		result[i], result[j] = result[j], result[i]
	}

	return result
}

// getWorkersAndShutdownOrder returns a copy of all workers and the shutdown order.
func (d *OrderedDaemon) getWorkersAndShutdownOrder() (map[string]*worker, []string) {
	d.lock.RLock()
	defer d.lock.RUnlock()

	workers := make(map[string]*worker)
	for k, v := range d.workers {
		workers[k] = v
	}

	shutdownOrderWorker := make([]string, len(d.shutdownOrderWorker))
	copy(shutdownOrderWorker, d.shutdownOrderWorker)

	return workers, shutdownOrderWorker
}

func (d *OrderedDaemon) runBackgroundWorker(name string, backgroundWorker WorkerFunc) {
	worker := d.workers[name]
	shutdownOrderWaitGroup := d.wgPerSameShutdownOrder[worker.shutdownOrder]
	shutdownOrderWaitGroup.Add(1)

	worker.running.Store(true)
	go func() {
		if d.logger != nil {
			d.logger.LogDebugf("Starting Background Worker: %s ...", name)
		}

		backgroundWorker(worker.ctx)

		// first we need to finish the waitgroup, otherwise stopWorkers could
		// already have acquired the lock and wait until all wait groups are done.
		shutdownOrderWaitGroup.Done()

		// now we can acquire the lock and cleanup the worker
		d.cleanupWorker(name)

		// only after cleanup is finished, we can unset the running flag,
		// otherwise there is a race condition between starting another worker with the same name
		// and a worker that is scheduled for cleanup.
		worker.running.Store(false)

		if d.logger != nil {
			d.logger.LogDebugf("Stopping Background Worker: %s ... done", name)
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

	exWorker, workerExistsAlready := d.workers[name]
	if workerExistsAlready {
		if !d.running.Load() {
			return ierrors.Wrapf(ErrDuplicateBackgroundWorker, "tried to overwrite existing background worker (%s)", name)
		}

		if exWorker.running.Load() {
			return ierrors.Wrapf(ErrExistingBackgroundWorkerStillRunning, "%s is still running", name)
		}

		// remove the existing worker from the shutdown order
		d.removeWorkerFromShutdownOrder(name)
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

	ctx, ctxCancel := context.WithCancel(context.Background())

	d.workers[name] = &worker{
		ctx:           ctx,
		ctxCancel:     ctxCancel,
		handler:       handler,
		shutdownOrder: shutdownOrder,
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

// DebugLogger allows to pass a logger to the daemon to issue log messages for debugging purposes.
func (d *OrderedDaemon) DebugLogger(logger log.Logger) {
	defaultDaemon.logger = logger
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
		d.running.Store(true)
		for name, worker := range d.workers {
			d.runBackgroundWorker(name, worker.handler)
		}
	}
}

// Run runs the daemon and then waits for the daemon to shutdown.
func (d *OrderedDaemon) Run() {
	d.Start()

	// wait until all wait groups for all shutdown orders are finished
	for _, wg := range d.waitGroupsForAllShutdownOrders() {
		if wg == nil {
			continue
		}
		wg.Wait()
	}
}

// returns all waitgroups of all existing shutdown orders or nil if none.
func (d *OrderedDaemon) waitGroupsForAllShutdownOrders() []*sync.WaitGroup {
	d.lock.RLock()
	defer d.lock.RUnlock()

	if len(d.wgPerSameShutdownOrder) == 0 {
		return nil
	}

	waitGroups := make([]*sync.WaitGroup, len(d.wgPerSameShutdownOrder))
	i := 0
	for _, wg := range d.wgPerSameShutdownOrder {
		waitGroups[i] = wg
		i++
	}

	return waitGroups
}

func (d *OrderedDaemon) shutdown() {
	if d.logger != nil {
		d.logger.LogDebugf("Shutting down ...")
	}

	d.stopped.Store(true)
	d.stoppedCtxCancel()
	if !d.IsRunning() {
		return
	}

	d.stopWorkers()
	d.running.Store(false)
	d.clear()
}

// stopWorkers stops all the workers of the daemon.
func (d *OrderedDaemon) stopWorkers() {
	workers, shutdownOrderWorker := d.getWorkersAndShutdownOrder()

	// stop all the workers
	if len(shutdownOrderWorker) > 0 {
		// initialize with the priority of the first worker
		prevPriority := workers[shutdownOrderWorker[0]].shutdownOrder
		for _, name := range shutdownOrderWorker {
			worker := workers[name]
			if !worker.running.Load() {
				worker.ctxCancel()

				continue
			}
			// if the current worker has a lower priority...
			if worker.shutdownOrder < prevPriority {
				// wait for every worker in the previous shutdown priority to terminate
				d.wgPerSameShutdownOrder[prevPriority].Wait()
				prevPriority = worker.shutdownOrder
			}
			if d.logger != nil {
				d.logger.LogDebugf("Stopping Background Worker: %s ...", name)
			}
			worker.ctxCancel()
		}
		// wait for the last priority to finish
		d.wgPerSameShutdownOrder[prevPriority].Wait()
	}
}

// cleanupWorker removes a finished worker from the workers pool and the shutdown order.
// Attention: this should only be called if the worker is already finished.
func (d *OrderedDaemon) cleanupWorker(name string) {
	d.lock.Lock()
	defer d.lock.Unlock()

	if d.IsStopped() {
		return
	}

	delete(d.workers, name)
	d.removeWorkerFromShutdownOrder(name)
}

// removeWorkerFromShutdownOrder removes the existing worker from the shutdown order.
// Attention: this should only be called if the worker is already finished.
func (d *OrderedDaemon) removeWorkerFromShutdownOrder(name string) {
	if d.shutdownOrderWorker == nil {
		return
	}

	for i, exName := range d.shutdownOrderWorker {
		if exName != name {
			continue
		}

		if i < len(d.shutdownOrderWorker)-1 {
			copy(d.shutdownOrderWorker[i:], d.shutdownOrderWorker[i+1:])
		}
		d.shutdownOrderWorker[len(d.shutdownOrderWorker)-1] = "" // mark for garbage collection with zero value
		d.shutdownOrderWorker = d.shutdownOrderWorker[:len(d.shutdownOrderWorker)-1]

		// we found the worker, no need to iterate further
		break
	}
}

// clear clears the daemon.
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
	return d.running.Load()
}

// IsStopped checks whether the daemon was stopped.
func (d *OrderedDaemon) IsStopped() bool {
	return d.stopped.Load()
}

// ContextStopped returns a context that is done when the deamon is stopped.
func (d *OrderedDaemon) ContextStopped() context.Context {
	return d.stoppedCtx
}
