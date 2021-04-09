package service

import (
	"encoding/json"
	"fmt"

	"golang.org/x/xerrors"

	pb "github.com/iotaledger/hive.go/autopeering/peer/service/proto"
	"google.golang.org/protobuf/proto"
)

// Record defines the mapping between a service ID and its tuple TypePort
// e.g., map[autopeering:&{TCP, 8000}]
type Record struct {
	m map[string]endpoint
}

// endpoint implements net.Addr
type endpoint struct {
	network string
	port    int
}

// Network returns the service's network name.
func (a endpoint) Network() string {
	return a.network
}

// Port returns the service's port number.
func (a endpoint) Port() int {
	return a.port
}

// String returns the service's address in string form.
func (a endpoint) String() string {
	return fmt.Sprintf("%d/%s", a.port, a.network)
}

// New initializes and returns an empty new Record
func New() *Record {
	return &Record{
		m: make(map[string]endpoint),
	}
}

// Get returns the network end point address of the service with the given name.
func (s *Record) Get(key Key) Endpoint {
	val, ok := s.m[string(key)]
	if !ok {
		return nil
	}
	return val
}

// CreateRecord creates a modifyable Record from the services.
func (s *Record) CreateRecord() *Record {
	result := New()
	for k, v := range s.m {
		result.m[k] = v
	}
	return result
}

// Update adds a new service to the map.
func (s *Record) Update(key Key, network string, port int) {
	s.m[string(key)] = endpoint{
		network: network,
		port:    port,
	}
}

// String returns a string representation of the service record.
func (s *Record) String() string {
	return fmt.Sprintf("%v", s.m)
}

// FromProto creates a Record from the provided protobuf struct.
func FromProto(in *pb.ServiceMap) (*Record, error) {
	m := in.GetMap()
	if m == nil {
		return nil, nil
	}

	services := New()
	for service, addr := range m {
		services.m[service] = endpoint{
			network: addr.GetNetwork(),
			port:    int(addr.GetPort()),
		}
	}
	return services, nil
}

// ToProto returns the corresponding protobuf struct.
func (s *Record) ToProto() *pb.ServiceMap {
	if len(s.m) == 0 {
		return nil
	}

	services := make(map[string]*pb.NetworkAddress, len(s.m))
	for service, addr := range s.m {
		services[service] = &pb.NetworkAddress{
			Network: addr.network,
			Port:    uint32(addr.port),
		}
	}

	return &pb.ServiceMap{
		Map: services,
	}
}

type endpointJSON struct {
	Network string `json:"network"`
	Port    int    `json:"port"`
}

// UnmarshalJSON deserializes JSON data into Record struct.
func (s *Record) UnmarshalJSON(b []byte) error {
	m := map[string]endpointJSON{}
	if err := json.Unmarshal(b, &m); err != nil {
		return xerrors.Errorf("failed to parse services map: %w", err)
	}
	services := New()
	for service, addr := range m {
		services.m[service] = endpoint{network: addr.Network, port: addr.Port}
	}
	*s = *services
	return nil
}

// MarshalJSON serializes Record struct into JSON data.
func (s *Record) MarshalJSON() ([]byte, error) {
	m := make(map[string]endpointJSON, len(s.m))
	for service, addr := range s.m {
		m[service] = endpointJSON{
			Network: addr.network,
			Port:    addr.port,
		}
	}
	return json.Marshal(m)
}

// Marshal serializes a given Peer (p) into a slice of bytes.
func (s *Record) Marshal() ([]byte, error) {
	return proto.Marshal(s.ToProto())
}

// Unmarshal de-serializes a given slice of bytes (data) into a Peer.
func Unmarshal(data []byte) (*Record, error) {
	s := &pb.ServiceMap{}
	if err := proto.Unmarshal(data, s); err != nil {
		return nil, err
	}
	return FromProto(s)
}
