package shutdown

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/iotaledger/hive.go/core/daemon"
	"github.com/iotaledger/hive.go/core/events"
	"github.com/iotaledger/hive.go/core/generics/options"
	"github.com/iotaledger/hive.go/core/logger"
)

// AppSelfShutdownCaller is used to signal a app self shutdown caused by an error.
func AppSelfShutdownCaller(handler interface{}, params ...interface{}) {
	handler.(func(reason string, critical bool))(params[0].(string), params[1].(bool))
}

type selfShutdownRequest struct {
	message  string
	critical bool
}

// Events holds Shutdown related events.
type Events struct {
	// Fired when a app self shutdown was caused by an error.
	AppSelfShutdown *events.Event
	// Fired when a clean shutdown was requested.
	AppShutdown *events.Event
}

// ShutdownHandler waits until a shutdown signal was received or the app tried to shutdown itself,
// and shuts down all processes gracefully.
//
//nolint:revive // better be explicit here
type ShutdownHandler struct {
	// the logger used to log events.
	*logger.WrappedLogger

	daemon          daemon.Daemon
	stopGracePeriod time.Duration

	gracefulStop    chan os.Signal
	appSelfShutdown chan selfShutdownRequest

	// Events are the events that are triggered by the ShutdownHandler.
	Events *Events
}

// WithStopGracePeriod defines the the maximum time to wait for background
// processes to finish during shutdown before terminating the app.
func WithStopGracePeriod(stopGracePeriod time.Duration) options.Option[ShutdownHandler] {
	return func(s *ShutdownHandler) {
		s.stopGracePeriod = stopGracePeriod
	}
}

// NewShutdownHandler creates a new shutdown handler.
func NewShutdownHandler(log *logger.Logger, daemon daemon.Daemon, opts ...options.Option[ShutdownHandler]) *ShutdownHandler {

	gs := options.Apply(&ShutdownHandler{
		WrappedLogger:   logger.NewWrappedLogger(log),
		daemon:          daemon,
		stopGracePeriod: 300 * time.Second,
		gracefulStop:    make(chan os.Signal, 1),
		appSelfShutdown: make(chan selfShutdownRequest),
		Events: &Events{
			AppSelfShutdown: events.NewEvent(AppSelfShutdownCaller),
			AppShutdown:     events.NewEvent(events.VoidCaller),
		},
	}, opts)

	signal.Notify(gs.gracefulStop, syscall.SIGTERM)
	signal.Notify(gs.gracefulStop, syscall.SIGINT)

	return gs
}

// SelfShutdown can be called in order to instruct the app to shutdown cleanly without receiving any interrupt signals.
func (gs *ShutdownHandler) SelfShutdown(msg string, critical bool) {
	select {
	case gs.appSelfShutdown <- selfShutdownRequest{message: msg, critical: critical}:
	default:
	}
}

// Run starts the ShutdownHandler go routine.
func (gs *ShutdownHandler) Run() {

	go func() {
		select {
		case <-gs.gracefulStop:
			gs.LogWarnf("Received shutdown request - waiting (max %d seconds) to finish processing ...", int(gs.stopGracePeriod.Seconds()))
			gs.Events.AppShutdown.Trigger()

		case selfShutdownReq := <-gs.appSelfShutdown:
			shutdownMsg := fmt.Sprintf("App self-shutdown: %s; waiting (max %d seconds) to finish processing ...", selfShutdownReq.message, int(gs.stopGracePeriod.Seconds()))
			if selfShutdownReq.critical {
				shutdownMsg = fmt.Sprintf("Critical %s", shutdownMsg)
			}
			gs.LogWarn(shutdownMsg)
			gs.Events.AppSelfShutdown.Trigger(selfShutdownReq.message, selfShutdownReq.critical)
		}

		go func() {
			ts := time.Now()
			for range time.Tick(1 * time.Second) {
				secondsSinceStart := int(time.Since(ts).Seconds())

				if secondsSinceStart <= int(gs.stopGracePeriod.Seconds()) {
					processList := ""
					runningBackgroundWorkers := gs.daemon.GetRunningBackgroundWorkers()
					if len(runningBackgroundWorkers) >= 1 {
						processList = "(" + strings.Join(runningBackgroundWorkers, ", ") + ") "
					}

					gs.LogWarnf("Received shutdown request - waiting (max %d seconds) to finish processing %s...", int(gs.stopGracePeriod.Seconds())-secondsSinceStart, processList)
				} else {
					gs.LogFatalAndExit("Background processes did not terminate in time! Forcing shutdown ...")
				}
			}
		}()

		gs.daemon.ShutdownAndWait()
	}()
}
