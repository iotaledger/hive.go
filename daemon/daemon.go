package daemon

import (
	"github.com/iotaledger/hive.go/syncutils"
	"github.com/pkg/errors"
	"sort"
	"sync"
	"time"
)

const (
	ShutdownPriorityLow    = 0
	ShutdownPriorityMedium = 1
	ShutdownPriorityHigh   = 2
)

var (
	ErrBackgroundWorkerAlreadyDefined = errors.New("background worker already defined")
)

// functions kept for backwards compatibility
var defaultDaemon = New()

// The ShutdownSignal held by the default daemon instance.
var ShutdownSignal = defaultDaemon.globalShutdownSignal

// GetRunningBackgroundWorkers gets the running background workers of the default daemon instance.
func GetRunningBackgroundWorkers() []string {
	return defaultDaemon.GetRunningBackgroundWorkers()
}

// BackgroundWorker adds a new background worker to the default daemon instance. Use priority
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

// Shutdown signals all background worker of the default deamon instance to shut down.
// This call doesn't await termination of the background workers.
func Shutdown(sleepBetweenPriorities ...time.Duration) {
	defaultDaemon.Shutdown(sleepBetweenPriorities...)
}

// Shutdown signals all background worker of the default deamon instance to shut down and
// then waits for their termination.
func ShutdownAndWait(sleepBetweenPriorities ...time.Duration) {
	defaultDaemon.ShutdownAndWait(sleepBetweenPriorities...)
}

// IsRunning checks whether the default daemon instance is running.
func IsRunning() bool {
	return defaultDaemon.IsRunning()
}

// New creates a new daemon instance.
func New() *Daemon {
	return &Daemon{
		running:              false,
		wg:                   sync.WaitGroup{},
		workers:              make(map[string]*worker),
		shutdownPriorities:   make([]string, 0),
		globalShutdownSignal: make(chan struct{}, 1),
		lock:                 syncutils.Mutex{},
	}
}

// Daemon is an orchestrator for background workers.
type Daemon struct {
	running              bool
	wg                   sync.WaitGroup
	workers              map[string]*worker
	shutdownPriorities   []string
	globalShutdownSignal chan struct{}
	lock                 syncutils.Mutex
}

// A function accepting its shutdown signal handler channel.
type WorkerFunc = func(shutdownSignal <-chan struct{})

type worker struct {
	priority       int
	handler        WorkerFunc
	running        bool
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
	d.wg.Add(1)

	go func() {
		d.lock.Lock()
		d.workers[name].running = true
		d.lock.Unlock()

		backgroundWorker(d.workers[name].shutdownSignal)

		d.lock.Lock()
		d.workers[name].running = false
		d.lock.Unlock()

		d.wg.Done()
	}()
}

// BackgroundWorker adds a new background worker to the daemon.
// Use priority to define in which shutdown order this particular
// background worker is shut down (higher = earlier).
func (d *Daemon) BackgroundWorker(name string, handler WorkerFunc, priority ...int) error {
	d.lock.Lock()

	_, has := d.workers[name]
	if has {
		return errors.Wrapf(ErrBackgroundWorkerAlreadyDefined, "%s is already defined", name)
	}

	var workerPriority int
	var shutdownSignal chan struct{}
	if len(priority) > 0 && priority[0] != 0 {
		workerPriority = priority[0]
		shutdownSignal = make(chan struct{}, 1)
	} else {
		shutdownSignal = d.globalShutdownSignal
	}

	d.workers[name] = &worker{
		priority:       workerPriority,
		handler:        handler,
		shutdownSignal: shutdownSignal,
	}

	// add to the shutdown sequence and order by priorities
	d.shutdownPriorities = append(d.shutdownPriorities, name)

	// must be done while holding the lock
	sort.Slice(d.shutdownPriorities, func(i, j int) bool {
		return d.workers[d.shutdownPriorities[i]].priority > d.workers[d.shutdownPriorities[j]].priority
	})

	if d.IsRunning() {
		d.runBackgroundWorker(name, handler)
	}

	d.lock.Unlock()
	return nil
}

// Start starts the daemon.
func (d *Daemon) Start() {
	if !d.running {
		d.lock.Lock()

		if !d.running {
			d.running = true

			Events.Run.Trigger()

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
	d.wg.Wait()
}

func (d *Daemon) shutdown(sleepBetweenPriorities ...time.Duration) {
	d.lock.Lock()
	defer d.lock.Unlock()

	if d.running {
		currentPriority := -1
		for _, name := range d.shutdownPriorities {
			worker := d.workers[name]
			if worker.priority == 0 {
				break
			}

			// sleep every time we come to a new priority
			if len(sleepBetweenPriorities) > 0 && (currentPriority == -1 || worker.priority < currentPriority) {
				if currentPriority != -1 {
					time.Sleep(sleepBetweenPriorities[0])
				}
				currentPriority = worker.priority
			}
			close(worker.shutdownSignal)
		}

		// global shutdown signal channel for all workers at priority 0
		close(d.globalShutdownSignal)

		d.running = false
		Events.Shutdown.Trigger()
	}
}

// Shutdown signals all background worker of the deamon shut down.
// This call doesn't await termination of the background workers.
func (d *Daemon) Shutdown(sleepBetweenPriorities ...time.Duration) {
	if d.running {
		d.shutdown(sleepBetweenPriorities...)
	}
}

// Shutdown signals all background worker of the deamon to shut down and
// then waits for their termination.
func (d *Daemon) ShutdownAndWait(sleepBetweenPriorities ...time.Duration) {
	if d.running {
		d.shutdown(sleepBetweenPriorities...)
	}
	d.wg.Wait()
}

// IsRunning checks whether the daemon is running.
func (d *Daemon) IsRunning() bool {
	return d.running
}
