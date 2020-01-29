package objectstorage

import (
	"fmt"
	"os"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/iotaledger/hive.go/platform"

	"github.com/iotaledger/hive.go/reflect"
)

// region interfaces ///////////////////////////////////////////////////////////////////////////////////////////////////

type LeakDetectionWrapper interface {
	CachedObject

	SetRetainCallStack(callStack *reflect.CallStack)
	GetRetainCallStack() *reflect.CallStack
	GetRetainTime() time.Time
	SetReleaseCallStack(callStack *reflect.CallStack)
	GetReleaseCallStack() *reflect.CallStack
	WasReleased() bool
}

// endregion ///////////////////////////////////////////////////////////////////////////////////////////////////////////

// region interface implementations ////////////////////////////////////////////////////////////////////////////////////

type LeakDetectionWrapperImpl struct {
	*CachedObjectImpl

	retainTime       time.Time
	released         int32
	retainCallStack  *reflect.CallStack
	releaseCallStack *reflect.CallStack
}

func (wrappedCachedObject *LeakDetectionWrapperImpl) Retain() CachedObject {
	baseCachedObject := wrappedCachedObject.CachedObjectImpl
	baseCachedObject.Retain()

	result := wrapCachedObject(baseCachedObject, 0).(*LeakDetectionWrapperImpl)
	result.GetRetainCallStack()

	return result
}

func (wrappedCachedObject *LeakDetectionWrapperImpl) Release() {
	if atomic.AddInt32(&(wrappedCachedObject.released), 1) != 1 {
		wrappedCachedObject.SetReleaseCallStack(reflect.GetExternalCallers("objectstorage", 0))

		reportCachedObjectClosedTooOften(wrappedCachedObject)
	} else {
		baseCachedObject := wrappedCachedObject.CachedObjectImpl

		// unregister identifier in debug list
		wrappedCachedObject.GetRetainCallStack()

		baseCachedObject.Release()
	}
}

func (wrappedCachedObject *LeakDetectionWrapperImpl) GetRetainTime() time.Time {
	return wrappedCachedObject.retainTime
}

func (wrappedCachedObject *LeakDetectionWrapperImpl) SetRetainCallStack(retainCallStack *reflect.CallStack) {
	wrappedCachedObject.retainCallStack = retainCallStack
	wrappedCachedObject.retainTime = time.Now()
}

func (wrappedCachedObject *LeakDetectionWrapperImpl) GetRetainCallStack() *reflect.CallStack {
	return wrappedCachedObject.retainCallStack
}

func (wrappedCachedObject *LeakDetectionWrapperImpl) SetReleaseCallStack(releaseCallStack *reflect.CallStack) {
	wrappedCachedObject.releaseCallStack = releaseCallStack
}

func (wrappedCachedObject *LeakDetectionWrapperImpl) GetReleaseCallStack() *reflect.CallStack {
	return wrappedCachedObject.releaseCallStack
}

func (wrappedCachedObject *LeakDetectionWrapperImpl) WasReleased() bool {
	return atomic.LoadInt32(&wrappedCachedObject.released) != 0
}

// endregion ///////////////////////////////////////////////////////////////////////////////////////////////////////////

// region public API ///////////////////////////////////////////////////////////////////////////////////////////////////

var (
	messageChan = make(chan interface{})
)

func init() {
	go func() {
		for {
			if message, isString := (<-messageChan).(string); isString {
				fmt.Println(message)
			} else {
				os.Exit(-1)
			}
		}
	}()
}

var LeakDetection = struct {
	WrapCachedObject                 func(cachedObject *CachedObjectImpl, skipCallerFrames int) CachedObject
	ReportCachedObjectClosedTooOften func(wrappedCachedObject LeakDetectionWrapper)
	MonitorCachedObjectReleased      func(wrappedCachedObject LeakDetectionWrapper, options *LeakDetectionOptions)
}{
	WrapCachedObject:                 wrapCachedObject,
	ReportCachedObjectClosedTooOften: reportCachedObjectClosedTooOften,
	MonitorCachedObjectReleased:      monitorCachedObjectReleased,
}

func wrapCachedObject(baseCachedObject *CachedObjectImpl, skipCallerFrames int) CachedObject {
	if baseCachedObject == nil {
		return nil
	}

	options := baseCachedObject.objectStorage.options

	if wrapCachedObject := options.leakDetectionWrapper; wrapCachedObject != nil {
		wrappedCachedObject := wrapCachedObject(baseCachedObject)
		wrappedCachedObject.SetRetainCallStack(reflect.GetExternalCallers("objectstorage", skipCallerFrames))

		monitorCachedObjectReleased(wrappedCachedObject, options.leakDetectionOptions)

		return wrappedCachedObject
	}

	return baseCachedObject
}

func reportCachedObjectClosedTooOften(wrappedCachedObject LeakDetectionWrapper) {
	retainCallStack := wrappedCachedObject.GetRetainCallStack()
	releaseCallStack := wrappedCachedObject.GetReleaseCallStack()

	messageChan <- "[objectstorage::leakkdetection] CachedObject released too often:" + platform.LineBreak +
		"\tretained: " + retainCallStack.ExternalEntryPoint() + platform.LineBreak +
		"\treleased: " + releaseCallStack.ExternalEntryPoint() + platform.LineBreak +
		platform.LineBreak +
		"\tretain call stack (full):" + platform.LineBreak +
		retainCallStack.String() + platform.LineBreak +
		"\trelease call stack (full):" + platform.LineBreak +
		releaseCallStack.String()

	messageChan <- nil
}

func monitorCachedObjectReleased(wrappedCachedObject LeakDetectionWrapper, options *LeakDetectionOptions) {
	time.AfterFunc(options.MaxConsumerHoldTime, func() {
		if !wrappedCachedObject.WasReleased() {
			messageChan <- "[objectstorage::leakkdetection] possible leak detected - CachedObject not released for more than " + strconv.Itoa(int(time.Since(wrappedCachedObject.GetRetainTime()).Seconds())) + " seconds:" + platform.LineBreak +
				"\tretained: " + wrappedCachedObject.GetRetainCallStack().ExternalEntryPoint() + platform.LineBreak +
				platform.LineBreak +
				"\tretain call stack (full):" + platform.LineBreak +
				wrappedCachedObject.GetRetainCallStack().String()

			monitorCachedObjectReleased(wrappedCachedObject, options)
		}
	})
}

// endregion ///////////////////////////////////////////////////////////////////////////////////////////////////////////

// region options //////////////////////////////////////////////////////////////////////////////////////////////////////

type LeakDetectionOptions struct {
	MaxSingleEntityConsumers int
	MaxGlobalEntityConsumers int
	MaxConsumerHoldTime      time.Duration
}

// endregion ///////////////////////////////////////////////////////////////////////////////////////////////////////////
