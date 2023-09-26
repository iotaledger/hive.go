package discover

import (
	"time"

	"github.com/izuc/zipp.foundation/autopeering/peer"
	"github.com/izuc/zipp.foundation/logger"
)

// Default values for the global parameters.
const (
	DefaultReverifyInterval = 10 * time.Second
	DefaultQueryInterval    = 60 * time.Second
	DefaultMaxManaged       = 1000
	DefaultMaxReplacements  = 10
)

var (
	reverifyInterval = DefaultReverifyInterval // time interval after which the next peer is reverified
	queryInterval    = DefaultQueryInterval    // time interval after which peers are queried for new peers
	maxManaged       = DefaultMaxManaged       // maximum number of peers that can be managed
	maxReplacements  = DefaultMaxReplacements  // maximum number of peers kept in the replacement list
)

type options struct {
	log         *logger.Logger // Logger
	masterPeers []*peer.Peer   // list of master peers used for bootstrapping
}

// Option modifies discovery related settings.
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

// MasterPeers sets the masterPeers to use.
func MasterPeers(masterPeers []*peer.Peer) Option {
	return optionFunc(func(opts *options) {
		opts.masterPeers = masterPeers
	})
}

// Parameters holds the parameters that can be configured.
type Parameters struct {
	ReverifyInterval time.Duration // time interval after which the next peer is reverified
	QueryInterval    time.Duration // time interval after which peers are queried for new peers
	MaxManaged       int           // maximum number of peers that can be managed
	MaxReplacements  int           // maximum number of peers kept in the replacement list
}

// SetParameters sets the global parameters for this package.
// This function cannot be used concurrently.
func SetParameters(param Parameters) {
	if param.ReverifyInterval > 0 {
		reverifyInterval = param.ReverifyInterval
	} else {
		reverifyInterval = DefaultReverifyInterval
	}
	if param.QueryInterval > 0 {
		queryInterval = param.QueryInterval
	} else {
		queryInterval = DefaultQueryInterval
	}
	if param.MaxManaged > 0 {
		maxManaged = param.MaxManaged
	} else {
		maxManaged = DefaultMaxManaged
	}
	if param.MaxReplacements > 0 {
		maxReplacements = param.MaxReplacements
	} else {
		maxReplacements = DefaultMaxReplacements
	}
}
