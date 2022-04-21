package event

import (
	"testing"
)

func BenchmarkEvent_Trigger(b *testing.B) {
	event := New[int]()

	event.Hook(NewClosure(func(param1 int) {
		// do nothing just get called
	}))

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		event.Trigger(4)
	}
}
