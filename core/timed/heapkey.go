package timed

import (
	"time"
)

// region HeapKey /////////////////////////////////////////////////////////////////////////////////////////////////

type HeapKey time.Time

func (t HeapKey) CompareTo(other HeapKey) int {
	if time.Time(t).Before(time.Time(other)) {
		return -1
	}
	if time.Time(t).After(time.Time(other)) {
		return 1
	}
	return 0
}

// endregion ///////////////////////////////////////////////////////////////////////////////////////////////////////////
