package servertest

import (
	"net"
)

var ipv4Loopback = net.ParseIP("127.0.0.1")

// NewConn crates a new UDP connection that can be used for server.Server in tests.
func NewConn() *net.UDPConn {
	conn, err := net.ListenUDP("udp", &net.UDPAddr{IP: ipv4Loopback, Port: 0})
	if err != nil {
		panic(err)
	}
	return conn
}
