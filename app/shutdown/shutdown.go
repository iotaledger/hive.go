package shutdown

import (
	"fmt"
	"os"
	"os/signal"
	"path"
	"strings"
	"syscall"
	"time"

	"github.com/izuc/zipp.foundation/app/daemon"
	"github.com/izuc/zipp.foundation/logger"
	"github.com/izuc/zipp.foundation/runtime/event"
	"github.com/izuc/zipp.foundation/runtime/ioutils"
	"github.com/izuc/zipp.foundation/runtime/options"
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
	AppSelfShutdown *event.Event2[string, bool]
	// Fired when a clean shutdown was requested.
	AppShutdown *event.Event
}

// ShutdownHandler waits until a shutdown signal was received or the app tried to shutdown itself,
// and shuts down all processes gracefully.
//
//nolint:revive // better be explicit here
type ShutdownHandler struct {
	// the logger used to log events.
	*logger.WrappedLogger

	daemon daemon.Daemon

	stopGracePeriod          time.Duration
	selfShutdownLogsEnabled  bool
	selfShutdownLogsFilePath string

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

// WithSelfShutdownLogsEnabled defines whether to store self-shutdown events to a log file.
func WithSelfShutdownLogsEnabled(selfShutdownLogsEnabled bool) options.Option[ShutdownHandler] {
	return func(s *ShutdownHandler) {
		s.selfShutdownLogsEnabled = selfShutdownLogsEnabled
	}
}

// WithSelfShutdownLogsFilePath defines the file path to the self-shutdown log.
func WithSelfShutdownLogsFilePath(selfShutdownLogsFilePath string) options.Option[ShutdownHandler] {
	return func(s *ShutdownHandler) {
		s.selfShutdownLogsFilePath = selfShutdownLogsFilePath
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
			AppSelfShutdown: event.New2[string, bool](),
			AppShutdown:     event.New(),
		},
	}, opts)

	signal.Notify(gs.gracefulStop, syscall.SIGTERM)
	signal.Notify(gs.gracefulStop, syscall.SIGINT)

	return gs
}

func (gs *ShutdownHandler) checkSelfShutdownLogsDirectory() error {
	shutdownLogsDirectory := path.Dir(gs.selfShutdownLogsFilePath)
	if shutdownLogsDirectory == "." {
		// no directory given
		return nil
	}

	if err := ioutils.CreateDirectory(shutdownLogsDirectory, 0o700); err != nil {
		return fmt.Errorf("creating self-shutdown logs directory (%s) failed: %w", shutdownLogsDirectory, err)

	}

	return nil
}

func (gs *ShutdownHandler) writeSelfShutdownLogFile(msg string, critical bool) {
	if gs.selfShutdownLogsEnabled {

		//nolint:nosnakecase // false positive
		f, err := os.OpenFile(gs.selfShutdownLogsFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			gs.LogWarnf("self-shutdown log can't be opened, error: %s", err.Error())

			return
		}
		defer f.Close()

		message := msg
		if critical {
			message += " (CRITICAL)"
		}

		if _, err := f.WriteString(fmt.Sprintf("%s: %s\n", time.Now().Format(time.RFC3339), message)); err != nil {
			gs.LogWarnf("self-shutdown log can't be written, error: %s", err.Error())
		}
	}
}

// SelfShutdown can be called in order to instruct the app to shutdown cleanly without receiving any interrupt signals.
func (gs *ShutdownHandler) SelfShutdown(msg string, critical bool) {
	select {
	case gs.appSelfShutdown <- selfShutdownRequest{message: msg, critical: critical}:
	default:
	}
}

// Run starts the ShutdownHandler go routine.
func (gs *ShutdownHandler) Run() error {

	if gs.selfShutdownLogsEnabled {
		if err := gs.checkSelfShutdownLogsDirectory(); err != nil {
			return err
		}
	}

	go func() {
		critical := false

		select {
		case <-gs.gracefulStop:
			gs.LogWarnf("Received shutdown request - waiting (max %d seconds) to finish processing ...", int(gs.stopGracePeriod.Seconds()))
			gs.Events.AppShutdown.Trigger()

		case selfShutdownReq := <-gs.appSelfShutdown:
			shutdownMsg := fmt.Sprintf("App self-shutdown: %s; waiting (max %d seconds) to finish processing ...", selfShutdownReq.message, int(gs.stopGracePeriod.Seconds()))
			if selfShutdownReq.critical {
				shutdownMsg = fmt.Sprintf("Critical %s", shutdownMsg)
				critical = true
			}
			gs.LogWarn(shutdownMsg)
			gs.writeSelfShutdownLogFile(selfShutdownReq.message, selfShutdownReq.critical)
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

		if critical {
			os.Exit(1)
		}
	}()

	return nil
}
