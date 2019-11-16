package logger_test

import (
	"github.com/iotaledger/hive.go/events"
	"github.com/iotaledger/hive.go/logger"
	"io/ioutil"
	"testing"
)

func TestLoggerEvents(t *testing.T) {
	// don't actually print anything
	logger.InjectWriters(ioutil.Discard)

	msg := "123"
	log := logger.NewLogger("app")

	var msgThroughEvent string
	logger.Events.AnyMsg.Attach(events.NewClosure(func(logLvl logger.LogLevel, prefix string, msgViaEvent string) {
		msgThroughEvent = msgViaEvent
	}))
	log.Info(msg)

	if msgThroughEvent != msg {
		t.Fatalf("expected message in event to be %s but was %s", msg, msgThroughEvent)
	}
}
