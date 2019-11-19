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
var defaultDaemon = NewDaemon()
var ShutdownSignal = defaultDaemon.globalShutdownSignal

func GetRunningBackgroundWorkers() []string {
	return defaultDaemon.GetRunningBackgroundWorkers()
}

func BackgroundWorker(name string, handler WorkerFunc, priority ...int) error {
	return defaultDaemon.BackgroundWorker(name, handler, priority...)
}

func Start() {
	defaultDaemon.Start()
}

func Run() {
	defaultDaemon.Run()
}

func Shutdown(sleepBetweenPriorities ...time.Duration) {
	defaultDaemon.Shutdown(sleepBetweenPriorities...)
}

func ShutdownAndWait(sleepBetweenPriorities ...time.Duration) {
	defaultDaemon.ShutdownAndWait(sleepBetweenPriorities...)
}

func IsRunning() bool {
	return defaultDaemon.IsRunning()
}

func NewDaemon() *Daemon {
	return &Daemon{
		running:              false,
		wg:                   sync.WaitGroup{},
		workers:              make(map[string]*worker),
		shutdownPriorities:   make([]string, 0),
		globalShutdownSignal: make(chan struct{}, 1),
		lock:                 syncutils.Mutex{},
	}
}

type Daemon struct {
	running              bool
	wg                   sync.WaitGroup
	workers              map[string]*worker
	shutdownPriorities   []string
	globalShutdownSignal chan struct{}
	lock                 syncutils.Mutex
}

type WorkerFunc = func(shutdownSignal <-chan struct{})

type worker struct {
	priority       int
	handler        WorkerFunc
	running        bool
	shutdownSignal chan struct{}
}

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

func (d *Daemon) BackgroundWorker(name string, handler WorkerFunc, priority ...int) error {
	d.lock.Lock()

	_, has := d.workers[name]
	if has {
		return errors.Wrapf(ErrBackgroundWorkerAlreadyDefined, "%s is already defined", name)
	}

	var workerPriority int
	var shutdownSignal chan struct{}
	if len(priority) > 0 {
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
			// sleep every time we come to a new priority
			if len(sleepBetweenPriorities) > 0 && (currentPriority == -1 || worker.priority < currentPriority) {
				if currentPriority != -1 {
					time.Sleep(sleepBetweenPriorities[0])
				}
				currentPriority = worker.priority
			}
			close(worker.shutdownSignal)
		}
		close(d.globalShutdownSignal)

		d.running = false
		Events.Shutdown.Trigger()
	}
}

func (d *Daemon) Shutdown(sleepBetweenPriorities ...time.Duration) {
	if d.running {
		d.shutdown(sleepBetweenPriorities...)
	}
}

func (d *Daemon) ShutdownAndWait(sleepBetweenPriorities ...time.Duration) {
	if d.running {
		d.shutdown(sleepBetweenPriorities...)
	}
	d.wg.Wait()
}

func (d *Daemon) IsRunning() bool {
	return d.running
}
