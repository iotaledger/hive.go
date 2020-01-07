package daemon

import (
	"sort"
	"sync"

	"github.com/iotaledger/hive.go/syncutils"
	"github.com/pkg/errors"
)

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

// Shutdown signals all background worker of the default daemon instance to shut down and
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
		running:                false,
		stopped:                false,
		workers:                make(map[string]*worker),
		shutdownOrderWorker:    make([]string, 0),
		wgPerSameShutdownOrder: make(map[int]*sync.WaitGroup),
		lock:                   syncutils.Mutex{},
	}
}

// OrderedDaemon is an orchestrator for background workers.
type OrderedDaemon struct {
	running                bool
	stopped                bool
	workers                map[string]*worker
	shutdownOrderWorker    []string
	wgPerSameShutdownOrder map[int]*sync.WaitGroup
	lock                   syncutils.Mutex
}

type worker struct {
	handler        WorkerFunc
	running        bool
	shutdownOrder  int
	shutdownSignal chan struct{}
}

// GetRunningBackgroundWorkers gets the running background workers.
func (d *OrderedDaemon) GetRunningBackgroundWorkers() []string {
	d.lock.Lock()
	defer d.lock.Unlock()

	result := make([]string, 0)
	for name, worker := range d.workers {
		if !worker.running {
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

	d.workers[name].running = true
	go func() {
		backgroundWorker(d.workers[name].shutdownSignal)
		d.workers[name].running = false
		shutdownOrderWaitGroup.Done()
	}()
}

// BackgroundWorker adds a new background worker to the daemon.
// Use order to define in which shutdown order this particular
// background worker is shut down (higher = earlier).
func (d *OrderedDaemon) BackgroundWorker(name string, handler WorkerFunc, order ...int) error {
	d.lock.Lock()
	defer d.lock.Unlock()

	if d.stopped {
		return ErrDaemonAlreadyStopped
	}

	exWorker, has := d.workers[name]
	if has {
		if exWorker.running {
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
		shutdownOrder:  shutdownOrder,
		handler:        handler,
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
	if d.stopped {
		return
	}

	if !d.running {
		d.lock.Lock()

		if !d.running {
			d.running = true
			for name, worker := range d.workers {
				d.runBackgroundWorker(name, worker.handler)
			}
		}

		d.lock.Unlock()
	}
}

// Run runs the daemon and then waits for the daemon to shutdown.
func (d *OrderedDaemon) Run() {
	d.Start()
	d.waitForLastPriority()
}

// waits on the lowest shutdownOrderWorker wait group
func (d *OrderedDaemon) waitForLastPriority() {
	if len(d.shutdownOrderWorker) == 0 {
		return
	}
	// find lowest shutdown order
	lowestShutdownPriorityWorker := d.workers[d.shutdownOrderWorker[len(d.shutdownOrderWorker)-1]]
	d.wgPerSameShutdownOrder[lowestShutdownPriorityWorker.shutdownOrder].Wait()
}

func (d *OrderedDaemon) shutdown() {
	d.lock.Lock()
	defer d.lock.Unlock()

	d.stopped = true
	if !d.running {
		return
	}

	currentPriority := -1
	for _, name := range d.shutdownOrderWorker {
		worker := d.workers[name]
		if !worker.running {
			// the worker's shutdown channel will be automatically garbage collected
			continue
		}
		if currentPriority == -1 || worker.shutdownOrder < currentPriority {
			if currentPriority != -1 {
				// wait for every worker in the same shutdown order to terminate
				d.wgPerSameShutdownOrder[currentPriority].Wait()
			}
			currentPriority = worker.shutdownOrder
		}

		close(worker.shutdownSignal)
	}

	// special case if we only had one order defined
	if currentPriority == -1 {
		currentPriority = 0
	}
	d.wgPerSameShutdownOrder[currentPriority].Wait()

	// clear
	d.running = false
	d.workers = make(map[string]*worker)
	d.shutdownOrderWorker = make([]string, 0)
	d.wgPerSameShutdownOrder = make(map[int]*sync.WaitGroup)
}

// Shutdown signals all background worker of the daemon shut down.
// This call doesn't await termination of the background workers.
func (d *OrderedDaemon) Shutdown() {
	go d.shutdown()
}

// Shutdown signals all background worker of the daemon to shut down and
// then waits for their termination.
func (d *OrderedDaemon) ShutdownAndWait() {
	d.shutdown()
}

// IsRunning checks whether the daemon is running.
func (d *OrderedDaemon) IsRunning() bool {
	return d.running
}

// IsStopped checks whether the daemon was stopped.
func (d *OrderedDaemon) IsStopped() bool {
	return d.stopped
}
