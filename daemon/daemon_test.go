package daemon_test

import (
	"log"
	"strconv"
	"testing"

	ordered_daemon "github.com/iotaledger/hive.go/daemon"
)

func TestStartShutdown(t *testing.T) {

	daemonA := ordered_daemon.New()

	var isShutdown, wasStarted bool
	if err := daemonA.BackgroundWorker("A", func(shutdownSignal <-chan struct{}) {
		wasStarted = true
		<-shutdownSignal
		isShutdown = true
	}); err != nil {
		t.Fatal(err)
	}

	daemonA.Start()
	daemonA.ShutdownAndWait()

	if !wasStarted {
		log.Fatalf("expected worker A to be started")
	}

	if !isShutdown {
		log.Fatalf("expected worker A to be shutdown")
	}
}

func TestShutdownOrder(t *testing.T) {

	daemonB := ordered_daemon.New()

	order := []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10}

	feedback := make(chan int, len(order))
	for _, o := range order {
		func(o int) {
			if err := daemonB.BackgroundWorker(strconv.Itoa(o), func(shutdownSignal <-chan struct{}) {
				<-shutdownSignal
				feedback <- o
			}, o); err != nil {
				t.Fatal(err)
			}
		}(o)
	}

	daemonB.Start()
	daemonB.ShutdownAndWait()

	for i := len(order) - 1; i >= 0; i-- {
		n := <-feedback
		if n != i {
			t.Fatalf("wrong shutdown sequence, expected %d but was %d", i, n)
		}
	}
}

func TestReRun(t *testing.T) {

	daemonC := ordered_daemon.New()

	terminate := make(chan struct{}, 1)
	feedback := make(chan struct{}, 1)

	if err := daemonC.BackgroundWorker("A", func(shutdownSignal <-chan struct{}) {
		<-terminate
		feedback <- struct{}{}
	}); err != nil {
		t.Fatal(err)
	}

	daemonC.Start()

	terminate <- struct{}{}
	<-feedback

	if err := daemonC.BackgroundWorker("A", func(shutdownSignal <-chan struct{}) {
		<-shutdownSignal
	}); err != nil {
		t.Fatal(err)
	}

	daemonC.ShutdownAndWait()
}
