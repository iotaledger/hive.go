package selection

import (
	"time"

	"github.com/iotaledger/hive.go/autopeering/peer"
	"github.com/iotaledger/hive.go/logger"
)

// Default values for the global parameters
const (
	DefaultInboundNeighborSize        = 4
	DefaultOutboundNeighborSize       = 4
	DefaultArrowLifetime              = 2 * time.Hour
	DefaultEpochDuration              = 1 * time.Hour
	DefaultOutboundUpdateInterval     = 1 * time.Second
	DefaultFullOutboundUpdateInterval = 1 * time.Minute
)

var (
	inboundNeighborSize        = DefaultInboundNeighborSize        // number of inbound neighbors
	outboundNeighborSize       = DefaultOutboundNeighborSize       // number of outbound neighbors
	arrowLifetime              = DefaultArrowLifetime              // lifetime of the arrow values
	outboundUpdateInterval     = DefaultOutboundUpdateInterval     // time after which out neighbors are updated
	fullOutboundUpdateInterval = DefaultFullOutboundUpdateInterval // time after which full out neighbors are updated
)

type options struct {
	log               *logger.Logger // Logger
	dropOnUpdate      bool           // set true to drop all neighbors when the arrow is updated
	neighborValidator Validator      // potential neighbor validator
}

// An Option configures the peer selection.
type Option interface {
	apply(*options)
}

// optionFunc wraps a func so it satisfies the Option interface.
type optionFunc func(*options)

func (f optionFunc) apply(opts *options) { f(opts) }

// Logger sets the logger
func Logger(log *logger.Logger) Option {
	return optionFunc(func(opts *options) {
		opts.log = log
	})
}

// DropOnUpdate sets the Option to drop all neighbors when the arrow is updated
func DropOnUpdate(dropOnUpdate bool) Option {
	return optionFunc(func(opts *options) {
		opts.dropOnUpdate = dropOnUpdate
	})
}

// NeighborValidator sets the potential neighbor validator
func NeighborValidator(neighborValidator Validator) Option {
	return optionFunc(func(opts *options) {
		opts.neighborValidator = neighborValidator
	})
}

// A Validator checks whether a peer is a valid neighbor
type Validator interface {
	IsValid(*peer.Peer) bool
}

// The ValidatorFunc type is an adapter to allow the use of ordinary functions as neighbor validators.
// If f is a function with the appropriate signature, ValidatorFunc(f) is a Validator that calls f.
type ValidatorFunc func(p *peer.Peer) bool

// IsValid calls f(p).
func (f ValidatorFunc) IsValid(p *peer.Peer) bool { return f(p) }

// Parameters holds the parameters that can be configured.
type Parameters struct {
	InboundNeighborSize        int           // number of inbound neighbors
	OutboundNeighborSize       int           // number of outbound neighbors
	ArRowLifetime              time.Duration // lifetime of the private and public local arrow
	OutboundUpdateInterval     time.Duration // time interval after which the outbound neighbors are checked
	FullOutboundUpdateInterval time.Duration // time after which the full outbound neighbors are updated
}

// SetParameters sets the global parameters for this package.
// This function cannot be used concurrently.
func SetParameters(param Parameters) {
	if param.InboundNeighborSize > 0 {
		inboundNeighborSize = param.InboundNeighborSize
	} else {
		inboundNeighborSize = DefaultInboundNeighborSize
	}
	if param.OutboundNeighborSize > 0 {
		outboundNeighborSize = param.OutboundNeighborSize
	} else {
		outboundNeighborSize = DefaultOutboundNeighborSize
	}
	if param.ArRowLifetime > 0 {
		arrowLifetime = param.ArRowLifetime
	} else {
		arrowLifetime = DefaultArrowLifetime
	}
	if param.OutboundUpdateInterval > 0 {
		outboundUpdateInterval = param.OutboundUpdateInterval
	} else {
		outboundUpdateInterval = DefaultOutboundUpdateInterval
	}
	if param.FullOutboundUpdateInterval > 0 {
		fullOutboundUpdateInterval = param.FullOutboundUpdateInterval
	} else {
		fullOutboundUpdateInterval = DefaultFullOutboundUpdateInterval
	}
}
