package debug

import (
	"fmt"
	"os"
	"os/signal"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
)

var (
	enabled      = false
	enabledMutex sync.RWMutex

	// DeadlockDetectionTimeout contains the duration to wait before assuming a deadlock.
	DeadlockDetectionTimeout = 5 * time.Second
)

// GetEnabled returns true if the debug mode is active.
func GetEnabled() bool {
	enabledMutex.RLock()
	defer enabledMutex.RUnlock()

	return enabled
}

// SetEnabled sets if the debug mode is active.
func SetEnabled(newEnabled bool) {
	enabledMutex.Lock()
	defer enabledMutex.Unlock()

	enabled = newEnabled
}

// FunctionName returns the name of the generic function pointer.
func FunctionName(functionPointer interface{}) string {
	return runtime.FuncForPC(reflect.ValueOf(functionPointer).Pointer()).Name()
}

// StackTrace returns a goroutine stack trace. If the allGoRoutines parameter is false, then it only returns the stack
// trace of the calling goroutine. It is possible to skip the first n frames of the stack trace.
func StackTrace(allGoRoutines bool, skipFrames int) string {
	buf := make([]byte, 1<<20)
	str := string(buf[:runtime.Stack(buf, allGoRoutines)])
	lines := strings.Split(strings.ReplaceAll(str, "\r\n", "\n"), "\n")

	trimmedLines := make([]string, 0)
	trimmedLines = append(trimmedLines, lines[:1]...)
	trimmedLines = append(trimmedLines, lines[1+(1+skipFrames)*2:]...)

	return strings.Join(trimmedLines, "\n")
}

// CallerStackTrace returns a formatted stack trace of the caller of this function.
func CallerStackTrace() (stackTrace string) {
	return strings.TrimSuffix("\tcalled by "+strings.ReplaceAll(StackTrace(false, 1), "\n", "\n\t\t"), "\t\t")
}

// ClosureStackTrace returns a formatted stack trace for the given function pointer.
func ClosureStackTrace(functionPointer interface{}) (stackTrace string) {
	return strings.TrimSuffix("\tclosure:\n\t\t"+FunctionName(functionPointer)+"\n\n\tcalled by "+strings.ReplaceAll(StackTrace(false, 1), "\n", "\n\t\t"), "\t\t")
}

// GoroutineID returns the ID of the current goroutine.
func GoroutineID() uint64 {
	buf := make([]byte, 1<<20)
	str := string(buf[:runtime.Stack(buf, false)])
	str = strings.TrimPrefix(str, "goroutine ")
	str, _, _ = strings.Cut(str, " ")

	goRoutineID, err := strconv.ParseUint(str, 10, 64)
	if err != nil {
		panic(err)
	}

	return goRoutineID
}

// DumpGoRoutinesOnShutdown dumps the stack traces of all goroutines on shutdown.
func DumpGoRoutinesOnShutdown() {
	go func() {
		sigs := make(chan os.Signal, 1)
		signal.Notify(sigs, syscall.SIGQUIT, syscall.SIGINT, syscall.SIGTERM)
		<-sigs
		fmt.Println(StackTrace(true, 0))
		os.Exit(1)
	}()
}
