package discover

import (
	"time"

	"github.com/iotaledger/hive.go/autopeering/peer"
	"github.com/iotaledger/hive.go/logger"
)

// Default values for the global parameters
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

// Options holds discovery related settings.
type Options struct {
	Log         *logger.Logger // Logger
	Version     uint32         // Protocol version
	MasterPeers []*peer.Peer   // list of master peers used for bootstrapping
}

type option func(*Options)

// Logger sets the logger
func Logger(log *logger.Logger) option {
	return func(args *Options) {
		args.Log = log
	}
}

// Version sets the VersionNumber of the protocol
func Version(version uint32) option {
	return func(args *Options) {
		args.Version = version
	}
}

// MasterPeers sets the masterPeers to use
func MasterPeers(masterPeers []*peer.Peer) option {
	return func(args *Options) {
		args.MasterPeers = masterPeers
	}
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
func SetParameter(param Parameters) {
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
