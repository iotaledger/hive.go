package ordered

import (
	"github.com/iotaledger/hive.go/syncutils"
	"github.com/pkg/errors"
	"sort"
	"sync"
)

var (
	ErrBackgroundWorkerAlreadyDefined = errors.New("background worker already defined")
)

// functions kept for backwards compatibility
var defaultDaemon = New()

// GetRunningBackgroundWorkers gets the running background workers of the default daemon instance.
func GetRunningBackgroundWorkers() []string {
	return defaultDaemon.GetRunningBackgroundWorkers()
}

// BackgroundWorker adds a new background worker to the default daemon instance. Use shutdownOrder
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

// New creates a new daemon instance.
func New() *Daemon {
	return &Daemon{
		running:                false,
		workers:                make(map[string]*worker),
		shutdownOrder:          make([]string, 0),
		wgPerSameShutdownOrder: make(map[int]*sync.WaitGroup),
		lock:                   syncutils.Mutex{},
	}
}

// Daemon is an orchestrator for background workers.
type Daemon struct {
	running                bool
	workers                map[string]*worker
	shutdownOrder          []string
	wgPerSameShutdownOrder map[int]*sync.WaitGroup
	lock                   syncutils.Mutex
}

// A function accepting its shutdown signal handler channel.
type WorkerFunc = func(shutdownSignal <-chan struct{})

type worker struct {
	handler        WorkerFunc
	running        bool
	shutdownOrder  int
	shutdownSignal chan struct{}
}

// GetRunningBackgroundWorkers gets the running background workers.
func (d *Daemon) GetRunningBackgroundWorkers() []string {
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

func (d *Daemon) runBackgroundWorker(name string, backgroundWorker WorkerFunc) {
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
func (d *Daemon) BackgroundWorker(name string, handler WorkerFunc, order ...int) error {
	d.lock.Lock()
	defer d.lock.Unlock()

	_, has := d.workers[name]
	if has {
		return errors.Wrapf(ErrBackgroundWorkerAlreadyDefined, "%s is already defined", name)
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
	d.shutdownOrder = append(d.shutdownOrder, name)

	// must be done while holding the lock
	sort.Slice(d.shutdownOrder, func(i, j int) bool {
		return d.workers[d.shutdownOrder[i]].shutdownOrder > d.workers[d.shutdownOrder[j]].shutdownOrder
	})

	if d.IsRunning() {
		d.runBackgroundWorker(name, handler)
	}

	return nil
}

// Start starts the daemon.
func (d *Daemon) Start() {
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
func (d *Daemon) Run() {
	d.Start()
	d.waitForLastPriority()
}

// waits on the lowest shutdownOrder wait group
func (d *Daemon) waitForLastPriority() {
	if len(d.shutdownOrder) == 0 {
		return
	}
	lowestShutdownPriorityWorker := d.workers[d.shutdownOrder[len(d.shutdownOrder)-1]]
	d.wgPerSameShutdownOrder[lowestShutdownPriorityWorker.shutdownOrder].Wait()
}

func (d *Daemon) shutdown() {
	d.lock.Lock()
	defer d.lock.Unlock()

	if !d.running {
		return
	}
	currentPriority := -1
	for _, name := range d.shutdownOrder {
		worker := d.workers[name]
		if currentPriority == -1 || worker.shutdownOrder < currentPriority {
			if currentPriority != -1 {
				// wait for every worker in the shutdownOrder
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
	d.running = false
}

// Shutdown signals all background worker of the daemon shut down.
// This call doesn't await termination of the background workers.
func (d *Daemon) Shutdown() {
	if d.running {
		go d.shutdown()
	}
}

// Shutdown signals all background worker of the daemon to shut down and
// then waits for their termination.
func (d *Daemon) ShutdownAndWait() {
	if d.running {
		d.shutdown()
	}
	d.waitForLastPriority()
}

// IsRunning checks whether the daemon is running.
func (d *Daemon) IsRunning() bool {
	return d.running
}
