package tcp

import "github.com/iotaledger/hive.go/v2/events"

type tcpServerEvents struct {
	Start    *events.Event
	Shutdown *events.Event
	Connect  *events.Event
	Error    *events.Event
}
