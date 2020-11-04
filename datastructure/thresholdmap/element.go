package thresholdmap

import (
	"github.com/iotaledger/hive.go/datastructure/redblacktree"
)

type Element struct {
	*redblacktree.Node
}
