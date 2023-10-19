package reflect

import (
	"fmt"
	"path"
	"runtime"
	"strconv"
)

type CallStack struct {
	frames             []runtime.Frame
	externalEntryPoint int
}

func (callStack *CallStack) ExternalEntryPoint() string {
	externalCallerFrame := callStack.frames[callStack.externalEntryPoint]

	return externalCallerFrame.File + ":" + strconv.Itoa(externalCallerFrame.Line)
}

func (callStack *CallStack) String() string {
	result := ""

	for _, caller := range callStack.frames {
		result += "\t\t" + path.Base(caller.Function) + "(...)\r\n\t\t\t" + caller.File + ":" + strconv.Itoa(caller.Line) + "\r\n"
	}

	return result
}

type Callers []runtime.Frame

func (callers Callers) Skip(n int) Callers {
	return callers[n:]
}

func (callers Callers) String() string {
	result := ""

	for _, caller := range callers {
		result += path.Base(caller.Function) + "(...)\r\n\t" + caller.File + ":" + strconv.Itoa(caller.Line)
	}

	return result
}

func GetExternalCallers(packageName string, skipCallers int) (callers *CallStack) {
	programCountersSize, programCountersCapacity, programCounters := 64, 64, make([]uintptr, 64)
	for programCountersSize == programCountersCapacity {
		programCounters = make([]uintptr, programCountersCapacity)
		programCountersSize = runtime.Callers(2, programCounters)

		programCountersCapacity *= 2
	}

	callStack := &CallStack{
		frames:             make([]runtime.Frame, 0),
		externalEntryPoint: 0,
	}

	if programCountersSize >= 1 {
		frames := runtime.CallersFrames(programCounters[:programCountersSize])
		frameCounter := 0
		for {
			frame, frameExists := frames.Next()
			if !frameExists {
				break
			}

			if frameCounter := len(callStack.frames); callStack.externalEntryPoint == 0 && path.Base(path.Dir(frame.File)) != packageName {
				callStack.externalEntryPoint = frameCounter + skipCallers
			}

			callStack.frames = append(callStack.frames, frame)

			frameCounter++
		}
	}

	return callStack
}

func GetCallers(skipCallers int) (callers Callers) {
	programCountersSize, programCountersCapacity, programCounters := 64, 64, make([]uintptr, 64)
	for programCountersSize == programCountersCapacity {
		programCounters = make([]uintptr, programCountersCapacity)
		programCountersSize = runtime.Callers(skipCallers, programCounters)

		programCountersCapacity *= 2
	}

	if programCountersSize >= 1 {
		frames := runtime.CallersFrames(programCounters[:programCountersSize])
		for {
			frame, frameExists := frames.Next()
			if !frameExists {
				break
			}

			fmt.Println(path.Base(path.Dir(frame.File)))
			callers = append(callers, frame)
		}
	}

	return
}

func GetCaller(skipFrames int) runtime.Frame {
	// We need the frame at index skipFrames+2, since we never want runtime.Callers and getFrame
	targetFrameIndex := skipFrames + 2

	// Set size to targetFrameIndex+2 to ensure we have room for one more caller than we need
	programCounters := make([]uintptr, targetFrameIndex+2)
	n := runtime.Callers(0, programCounters)

	frame := runtime.Frame{Function: "unknown"}
	if n > 0 {
		frames := runtime.CallersFrames(programCounters[:n])
		for more, frameIndex := true, 0; more && frameIndex <= targetFrameIndex; frameIndex++ {
			var frameCandidate runtime.Frame
			frameCandidate, more = frames.Next()
			if frameIndex == targetFrameIndex {
				frame = frameCandidate
			}
		}
	}

	return frame
}
