package confirmation

const (
	// Pending is the default confirmation state (undecided).
	Pending State = iota

	// Rejected is the state for rejected entities.
	Rejected

	// Accepted is the state for accepted entities.
	Accepted

	// Confirmed is the state for confirmed entities.
	Confirmed
)

// State is the confirmation state of an entity.
type State uint8

// String returns a human-readable representation of the State.
func (s State) String() (humanReadable string) {
	switch s {
	case Pending:
		return "Pending"
	case Rejected:
		return "Rejected"
	case Accepted:
		return "Accepted"
	case Confirmed:
		return "Confirmed"
	default:
		return "Unknown"
	}
}
