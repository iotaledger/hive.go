package algorithm

import (
	"container/list"

	"github.com/iotaledger/hive.go/datastructure/set"
)

// dagElementID is a type alias representing the identifier for elements in a DAG.
type dagElementID = interface{}

// dagElement is a type alias representing the elements in a DAG.
type dagElement = interface{}

// WalkDAG is a generic walker that executes a custom callback for every visited element in a DAG, starting from
// the given entry points. The callback should return the dagElementIDs to be visited next. It accepts an optional
// boolean parameter which can be set to true if a dagElement should be visited more than once following different
// paths.
func WalkDAG(callback func(element dagElement) (nextElementIDsToVisit []dagElementID), entryPoints []dagElementID, revisitElements ...bool) {
	if len(entryPoints) == 0 {
		panic("you need to provide at least one entry point")
	}

	stack := list.New()
	for _, elementID := range entryPoints {
		stack.PushBack(elementID)
	}

	visitedElementIDs := set.New()
	revisit := len(revisitElements) != 0 && revisitElements[0]

	for stack.Len() > 0 {
		firstElement := stack.Front()
		stack.Remove(firstElement)

		elementID := firstElement.Value
		if !revisit && !visitedElementIDs.Add(elementID) {
			continue
		}

		for _, nextElementID := range callback(elementID) {
			stack.PushBack(nextElementID)
		}
	}
}
