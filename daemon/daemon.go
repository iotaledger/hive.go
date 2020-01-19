package daemon

import (
	"sort"
	"sync"

	"github.com/iotaledger/hive.go/syncutils"
	"github.com/iotaledger/hive.go/typeutils"
	"github.com/pkg/errors"
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

// New creates a new daemon instance.
func New() *OrderedDaemon {
	return &OrderedDaemon{
		running:                typeutils.NewAtomicBool(),
		stopped:                typeutils.NewAtomicBool(),
		workers:                make(map[string]*worker),
		shutdownOrderWorker:    make([]string, 0),
		wgPerSameShutdownOrder: make(map[int]*sync.WaitGroup),
	}
}

// OrderedDaemon is an orchestrator for background workers.
type OrderedDaemon struct {
	running                *typeutils.AtomicBool
	stopped                *typeutils.AtomicBool
	workers                map[string]*worker
	shutdownOrderWorker    []string
	wgPerSameShutdownOrder map[int]*sync.WaitGroup
	lock                   syncutils.Mutex
}

type worker struct {
	handler        WorkerFunc
	running        *typeutils.AtomicBool
	shutdownOrder  int
	shutdownSignal chan struct{}
}

// GetRunningBackgroundWorkers gets the running background workers.
func (d *OrderedDaemon) GetRunningBackgroundWorkers() []string {
	d.lock.Lock()
	defer d.lock.Unlock()

	result := make([]string, 0)
	for name, worker := range d.workers {
		if !worker.running.IsSet() {
			continue
		}
		result = append(result, name)
	}

	return result
}

func (d *OrderedDaemon) runBackgroundWorker(name string, backgroundWorker WorkerFunc) {
	worker := d.workers[name]
	shutdownOrderWaitGroup := d.wgPerSameShutdownOrder[worker.shutdownOrder]
	shutdownOrderWaitGroup.Add(1)

	worker.running.Set()
	go func() {
		backgroundWorker(worker.shutdownSignal)
		worker.running.UnSet()
		shutdownOrderWaitGroup.Done()
	}()
}

// BackgroundWorker adds a new background worker to the daemon.
// Use order to define in which shutdown order this particular
// background worker is shut down (higher = earlier).
func (d *OrderedDaemon) BackgroundWorker(name string, handler WorkerFunc, order ...int) error {
	d.lock.Lock()
	defer d.lock.Unlock()

	if d.IsStopped() {
		return ErrDaemonAlreadyStopped
	}

	exWorker, has := d.workers[name]
	if has {
		if exWorker.running.IsSet() {
			return errors.Wrapf(ErrExistingBackgroundWorkerStillRunning, "%s is still running", name)
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

// Start starts the daemon.
func (d *OrderedDaemon) Start() {
	d.lock.Lock()
	defer d.lock.Unlock()

	// do not allow restarts
	if d.IsStopped() {
		return
	}

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
	d.lock.Lock()
	defer d.lock.Unlock()

	d.stopped.Set()
	if !d.IsRunning() {
		return
	}

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
			close(worker.shutdownSignal)
		}
		// wait for the last priority to finish
		d.wgPerSameShutdownOrder[prevPriority].Wait()
	}

	// clear
	d.running.UnSet()
	d.workers = nil
	d.shutdownOrderWorker = nil
	d.wgPerSameShutdownOrder = nil
}

// Shutdown signals all background worker of the daemon shut down.
// This call doesn't await termination of the background workers.
func (d *OrderedDaemon) Shutdown() {
	go d.shutdown()
}

// ShutdownAndWait signals all background worker of the daemon to shut down and then waits for their termination.
func (d *OrderedDaemon) ShutdownAndWait() {
	d.shutdown()
}

// IsRunning checks whether the daemon is running.
func (d *OrderedDaemon) IsRunning() bool {
	return d.running.IsSet()
}

// IsStopped checks whether the daemon was stopped.
func (d *OrderedDaemon) IsStopped() bool {
	return d.stopped.IsSet()
}
