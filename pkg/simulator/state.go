package simulator

type state int

const (
	StateOffline state = iota
	StateRunning
	StateCrashed
)

func (s state) String() string {
	switch s {
	case StateOffline:
		return "offline"
	case StateRunning:
		return "running"
	case StateCrashed:
		return "crashed"
	default:
		return "unknown"
	}
}
