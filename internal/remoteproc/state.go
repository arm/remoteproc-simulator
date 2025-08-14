package remoteproc

type State int

const (
	StateOffline State = iota
	StateRunning
	StateCrashed
)

func (s State) String() string {
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
