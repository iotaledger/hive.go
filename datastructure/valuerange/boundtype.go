package valuerange

// BoundType indicates whether an EndPoint of some ValueRange is contained in the ValueRange itself ("closed") or not
// ("open"). If a range is unbounded on a side, it is neither open nor closed on that side; the bound simply does not
// exist.
type BoundType uint8

const (
	// BoundTypeOpen indicates that the EndPoint value is considered part of the ValueRange ("inclusive").
	BoundTypeOpen BoundType = iota

	// BoundTypeClosed indicates that the EndPoint value is not considered part of the ValueRange ("exclusive").
	BoundTypeClosed
)
