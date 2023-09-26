package selection

import (
	"time"

	"github.com/izuc/zipp.foundation/autopeering/mana"
	"github.com/izuc/zipp.foundation/autopeering/peer"
	"github.com/izuc/zipp.foundation/logger"
)

// Default values for the global parameters.
const (
	DefaultInboundNeighborSize        = 4
	DefaultOutboundNeighborSize       = 4
	DefaultSaltLifetime               = 2 * time.Hour
	DefaultOutboundUpdateInterval     = 1 * time.Second
	DefaultFullOutboundUpdateInterval = 1 * time.Minute
)

var (
	inboundNeighborSize        = DefaultInboundNeighborSize        // number of inbound neighbors
	outboundNeighborSize       = DefaultOutboundNeighborSize       // number of outbound neighbors
	saltLifetime               = DefaultSaltLifetime               // lifetime of the private and public local salt
	outboundUpdateInterval     = DefaultOutboundUpdateInterval     // time after which out neighbors are updated
	fullOutboundUpdateInterval = DefaultFullOutboundUpdateInterval // time after which full out neighbors are updated
)

type options struct {
	log                   *logger.Logger // Logger
	dropOnUpdate          bool           // set true to drop all neighbors when the salt is updated
	neighborValidator     Validator      // potential neighbor validator
	useMana               bool
	manaFunc              mana.Func
	r                     int
	ro                    float64
	neighborBlockDuration time.Duration
	neighborSkipTimeout   time.Duration
}

// An Option configures the peer selection.
type Option interface {
	apply(*options)
}

// optionFunc wraps a func so it satisfies the Option interface.
type optionFunc func(*options)

func (f optionFunc) apply(opts *options) { f(opts) }

// Logger sets the logger.
func Logger(log *logger.Logger) Option {
	return optionFunc(func(opts *options) {
		opts.log = log
	})
}

// DropOnUpdate sets the Option to drop all neighbors when the salt is updated.
func DropOnUpdate(dropOnUpdate bool) Option {
	return optionFunc(func(opts *options) {
		opts.dropOnUpdate = dropOnUpdate
	})
}

// NeighborValidator sets the potential neighbor validator.
func NeighborValidator(neighborValidator Validator) Option {
	return optionFunc(func(opts *options) {
		opts.neighborValidator = neighborValidator
	})
}

// UseMana sets the Option to use mana.
func UseMana(useMana bool) Option {
	return optionFunc(func(opts *options) {
		opts.useMana = useMana
	})
}

// ManaFunc sets the Option of the mana function to use.
func ManaFunc(manaFunc mana.Func) Option {
	return optionFunc(func(opts *options) {
		opts.manaFunc = manaFunc
	})
}

// R sets the Option for R.
func R(r int) Option {
	return optionFunc(func(opts *options) {
		opts.r = r
	})
}

// Ro sets the Option for Ro.
func Ro(ro float64) Option {
	return optionFunc(func(opts *options) {
		opts.ro = ro
	})
}

// NeighborBlockDuration sets the amount of time a peer should remain in the blocklist.
func NeighborBlockDuration(blockDuration time.Duration) Option {
	return optionFunc(func(opts *options) {
		opts.neighborBlockDuration = blockDuration
	})
}

// NeighborSkipTimeout sets the amount of time for which we should skip the peer
// and don't try to connect with it after any problem encountered with that peer.
func NeighborSkipTimeout(skipTimeout time.Duration) Option {
	return optionFunc(func(opts *options) {
		opts.neighborSkipTimeout = skipTimeout
	})
}

// A Validator checks whether a peer is a valid neighbor.
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
	SaltLifetime               time.Duration // lifetime of the private and public local salt
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
	if param.SaltLifetime > 0 {
		saltLifetime = param.SaltLifetime
	} else {
		saltLifetime = DefaultSaltLifetime
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
