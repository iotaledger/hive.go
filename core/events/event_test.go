package events

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"
)

func BenchmarkEvent_Trigger(b *testing.B) {
	event := NewEvent(intStringCaller)

	event.Hook(NewClosure(func(param1 int, param2 string) {
		// do nothing just get called
	}))

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		event.Trigger(4, "test")
	}
}

// define how the event converts the generic parameters to the typed params - ugly but go has no generics :(.
func intStringCaller(handler interface{}, params ...interface{}) {
	handler.(func(int, string))(params[0].(int), params[1].(string))
}

func Test_ExampleEvent(t *testing.T) {
	// create event object (usually exposed through a public struct that holds all the different event types)
	event := NewEvent(intStringCaller)

	triggerCountClosure1 := 0
	triggerCountClosure2 := 0

	// we have to wrap a function in a closure to make it identifiable
	closure1 := NewClosure(func(param1 int, param2 string) {
		fmt.Println("#1 " + param2 + ": " + strconv.Itoa(param1))
		triggerCountClosure1++
	})

	// multiple subscribers can hook to an event (closures can be inlined)
	event.Hook(closure1)
	event.Hook(NewClosure(func(param1 int, param2 string) {
		fmt.Println("#2 " + param2 + ": " + strconv.Itoa(param1))
		triggerCountClosure2++
	}))

	// trigger the event
	event.Trigger(1, "Hello World")

	require.Equal(t, 1, triggerCountClosure1)
	require.Equal(t, 1, triggerCountClosure2)

	// unsubscribe the first closure and trigger again
	event.Detach(closure1)
	event.Trigger(1, "Hello World")

	require.Equal(t, 1, triggerCountClosure1)
	require.Equal(t, 2, triggerCountClosure2)

	// Unordered output: #1 Hello World: 1
	// #2 Hello World: 1
	// #2 Hello World: 1
}

func Test_ExampleEvent_MaxTriggerCount(t *testing.T) {
	// create event object (usually exposed through a public struct that holds all the different event types)
	event := NewEvent(intStringCaller)

	triggerCountClosure1 := 0
	triggerCountClosure2 := 0
	triggerCountClosure3 := 0

	// we have to wrap a function in a closure to make it identifiable
	closure1 := NewClosure(func(param1 int, param2 string) {
		fmt.Println("#1 " + param2 + ": " + strconv.Itoa(param1))
		triggerCountClosure1++
	})

	// we have to wrap a function in a closure to make it identifiable
	closure2 := NewClosure(func(param1 int, param2 string) {
		fmt.Println("#2 " + param2 + ": " + strconv.Itoa(param1))
		triggerCountClosure2++
	})

	// multiple subscribers can hook to an event (closures can be inlined)
	event.Hook(closure1)
	event.Hook(closure2, 1)
	event.Hook(NewClosure(func(param1 int, param2 string) {
		fmt.Println("#3 " + param2 + ": " + strconv.Itoa(param1))
		triggerCountClosure3++
	}), 3)

	// trigger the event
	for i := 0; i < 10; i++ {
		event.Trigger(1, "Hello World", 2)
	}

	require.Equal(t, 10, triggerCountClosure1)
	require.Equal(t, 1, triggerCountClosure2)
	require.Equal(t, 3, triggerCountClosure3)

	// unsubscribe the first closure and trigger again
	event.Detach(closure1)

	event.Trigger(1, "Hello World", 2)

	require.Equal(t, 10, triggerCountClosure1)
	require.Equal(t, 1, triggerCountClosure2)
	require.Equal(t, 3, triggerCountClosure3)
}
