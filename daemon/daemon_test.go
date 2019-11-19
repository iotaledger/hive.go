package daemon_test

import (
	"github.com/iotaledger/hive.go/daemon"
	"log"
	"testing"
	"time"
)

func TestStartShutdown(t *testing.T) {

	daemonA := daemon.NewDaemon()

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

func TestShutdownPriority(t *testing.T) {

	daemonB := daemon.NewDaemon()

	const highShutdownPriorityWorker = "highShutdownPriorityWorker"
	const lowShutdownPriorityWorker = "lowShutdownPriorityWorker"

	feedback := make(chan string, 2)
	if err := daemonB.BackgroundWorker(highShutdownPriorityWorker, func(shutdownSignal <-chan struct{}) {
		<-shutdownSignal
		feedback <- highShutdownPriorityWorker
	}, daemon.ShutdownPriorityHigh); err != nil {
		t.Fatal(err)
	}

	if err := daemonB.BackgroundWorker(lowShutdownPriorityWorker, func(shutdownSignal <-chan struct{}) {
		<-shutdownSignal
		// fake slowness of shutdown so we get meaningful results
		// as otherwise both workers are ended roughly at the same time
		<-time.After(time.Duration(100) * time.Millisecond)
		feedback <- lowShutdownPriorityWorker
	}, daemon.ShutdownPriorityLow); err != nil {
		t.Fatal(err)
	}

	daemonB.Start()
	daemonB.ShutdownAndWait(time.Second)

	if <-feedback != highShutdownPriorityWorker {
		t.Fatalf("expected worker %s to be shutdown before worker %s", highShutdownPriorityWorker, lowShutdownPriorityWorker)
	}

}
