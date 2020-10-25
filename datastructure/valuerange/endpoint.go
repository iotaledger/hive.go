package valuerange

// EndPoint contains information about where ValueRanges start and end. It combines a threshold value with a BoundType.
type EndPoint struct {
	value     Value
	boundType BoundType
}
