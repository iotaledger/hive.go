package iputils

import (
	"fmt"
	"net"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

type IP struct {
	net.IP
}

func (ip *IP) IsIPv6() bool {
	return len(ip.To4()) != net.IPv4len
}

func (ip *IP) ToString() string {
	if ip.IsIPv6() {
		return "[" + ip.String() + "]"
	}

	return ip.String()
}

type IPAddresses struct {
	IPs map[*IP]struct{}
}

func NewIPAddresses() *IPAddresses {
	return &IPAddresses{IPs: make(map[*IP]struct{})}
}

func (ips *IPAddresses) Union(other *IPAddresses) *IPAddresses {
	union := NewIPAddresses()
	set := map[string]struct{}{}
	for ip := range ips.IPs {
		union.Add(ip)
		set[ip.String()] = struct{}{}
	}
	for ip := range other.IPs {
		if _, ok := set[ip.String()]; !ok {
			union.Add(ip)
		}
	}
	return union
}

func (ips *IPAddresses) GetPreferredAddress(preferIPv6 bool) *IP {
	if !preferIPv6 {
		for ip := range ips.IPs {
			if !ip.IsIPv6() {
				return ip
			}
		}
	} else {
		for ip := range ips.IPs {
			if ip.IsIPv6() {
				return ip
			}
		}
	}
	// it's a map/set
	for ip := range ips.IPs {
		return ip
	}
	return nil
}

func (ips *IPAddresses) Add(ip *IP) {
	ips.IPs[ip] = struct{}{}
}

func (ips *IPAddresses) Remove(ip *IP) {
	delete(ips.IPs, ip)
}

func (ips *IPAddresses) Len() int {
	return len(ips.IPs)
}

// OriginAddress represents a tuple of a IP or hostname, port and IPv6 preference
type OriginAddress struct {
	Addr       string
	Port       uint16
	PreferIPv6 bool
}

func (ra *OriginAddress) String() string {
	return fmt.Sprintf("%s:%d", ra.Addr, ra.Port)
}

var ErrOriginAddrInvalidAddrChunk = errors.New("invalid address chunk in origin address")
var ErrOriginAddrInvalidPort = errors.New("invalid port in origin address")

func ParseOriginAddress(s string) (*OriginAddress, error) {
	addressChunks := strings.Split(s, ":")
	if len(addressChunks) <= 1 {
		return nil, ErrOriginAddrInvalidAddrChunk
	}

	addr := strings.Join(addressChunks[:len(addressChunks)-1], ":")
	portInt, err := strconv.Atoi(addressChunks[len(addressChunks)-1])
	if err != nil {
		return nil, ErrOriginAddrInvalidPort
	}
	port := uint16(portInt)
	return &OriginAddress{Addr: addr, Port: port}, nil
}
