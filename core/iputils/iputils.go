package iputils

import (
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"
)

func IsIPv6(ip net.IP) bool {
	return len(ip.To4()) != net.IPv4len
}

func IPToString(ip net.IP) string {
	if IsIPv6(ip) {
		return "[" + ip.String() + "]"
	}

	return ip.String()
}

type IPAddresses struct {
	IPs map[*net.IP]struct{}
}

func NewIPAddresses() *IPAddresses {
	return &IPAddresses{IPs: make(map[*net.IP]struct{})}
}

func (ips *IPAddresses) Union(other *IPAddresses) *IPAddresses {
	union := NewIPAddresses()
	set := map[string]struct{}{}
	for ip := range ips.IPs {
		union.Add(*ip)
		set[ip.String()] = struct{}{}
	}
	for ip := range other.IPs {
		if _, ok := set[ip.String()]; !ok {
			union.Add(*ip)
		}
	}
	return union
}

func (ips *IPAddresses) GetPreferredAddress(preferIPv6 bool) net.IP {
	if !preferIPv6 {
		for ip := range ips.IPs {
			if !IsIPv6(*ip) {
				return *ip
			}
		}
	} else {
		for ip := range ips.IPs {
			if IsIPv6(*ip) {
				return *ip
			}
		}
	}
	// it's a map/set
	for ip := range ips.IPs {
		return *ip
	}
	return nil
}

func (ips *IPAddresses) Add(ip net.IP) {
	ips.IPs[&ip] = struct{}{}
}

func (ips *IPAddresses) Remove(ip net.IP) {
	delete(ips.IPs, &ip)
}

func (ips *IPAddresses) Len() int {
	return len(ips.IPs)
}

// OriginAddress represents a tuple of a IP or hostname, port, alias and IPv6 preference
type OriginAddress struct {
	Addr       string
	Port       uint16
	Alias      string
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

var ErrInvalidIPAddressOrHost = errors.New("invalid IP address or hostname")
var ErrNoIPAddressesFound = errors.New("could not resolve any IP address")

// GetIPAddressesFromHost returns all resolvable IP addresses (*IPAddresses) from a host.
// If it is an IP address this IP address will be returned as *IPAddresses
func GetIPAddressesFromHost(hostname string) (*IPAddresses, error) {
	ipAddresses := NewIPAddresses()

	// Check if it's an IPv6 address
	if strings.Contains(hostname, ":") {
		hostname = strings.ReplaceAll(hostname, "[", "")
		hostname = strings.ReplaceAll(hostname, "]", "")
		ip := net.ParseIP(hostname)
		if ip == nil {
			return nil, ErrInvalidIPAddressOrHost
		}

		ipAddresses.Add(ip)
		return ipAddresses, nil

	}

	// Check if it's an IPv4 address
	if ip := net.ParseIP(hostname); ip != nil {
		ipAddresses.Add(ip)
		return ipAddresses, nil
	}

	// If it's no IP addr, resolve them
	ipAddr, err := net.LookupHost(hostname)
	if err != nil {
		return nil, fmt.Errorf("%w: couldn't lookup IPs for %s", err, hostname)
	}

	if len(ipAddr) == 0 {
		return nil, fmt.Errorf("no IPs found for %s", hostname)
	}

	for _, addr := range ipAddr {
		ip := net.ParseIP(addr)
		if ip == nil {
			return nil, ErrInvalidIPAddressOrHost
		}
		ipAddresses.Add(ip)
	}

	if ipAddresses.Len() == 0 {
		return nil, ErrNoIPAddressesFound
	}

	return ipAddresses, nil
}
