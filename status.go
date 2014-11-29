package userconfig

type State string

const (
	StateFailed   State = "failed"
	StateDown     State = "down"
	StateStarting State = "starting"
	StateStopping State = "stopping"
	StateUp       State = "up"
	StateUnknown  State = "unknown"
)

func (s *State) String() string {
	return string(*s)
}

type SubState string

const (
	SubStartingWaitingOnDependency SubState = "waiting-on-dependency"
	SubStartingPreparingVolume              = "preparing-volume"
	SubStartingFetchingImage                = "fetching-image"
	SubStartingRegisteringService           = "registering-service"
	SubStartingRegisteringDomain            = "registering-domain"

	SubUnknown = "unknown"
)

func (s *SubState) String() string {
	return string(*s)
}

type Status struct {
	State State    `json:"state"`
	Sub   SubState `json:"sub"`
}

// AggregateStatus returns the 'higher' of the two status,
// given the following order:
//  ok < starting < stopping < down < failed
// If any of two state is unknown, the other one is returned
func AggregateState(status1, status2 State) State {
	if status1 == StateUnknown {
		return status2
	}
	if status2 == StateUnknown {
		return status1
	}

	// Hierarchy checks
	if status1 == StateFailed || status2 == StateFailed {
		return StateFailed
	}
	if status1 == StateDown || status2 == StateDown {
		return StateDown
	}
	if status1 == StateStopping || status2 == StateStopping {
		return StateStopping
	}
	if status1 == StateStarting || status2 == StateStarting {
		return StateStarting
	}
	return StateUp
}

// Active means starting or up.
func IsStatusActive(status Status) bool {
	return IsStateActive(status.State)
}

func IsStateActive(state State) bool {
	return state == StateStarting || state == StateUp
}

// IsStatusFinal returns whether the given status is a final status and should
// not change upon itself. E.g. A unit with a StateStarting will after some
// time either switch to StateUp or StateFailed, thus the state is not final.
// Final states are StateUp, StateDown or StateFailed.
func IsStatusFinal(status Status) bool {
	return IsStateFinal(status.State)
}
func IsStateFinal(state State) bool {
	return state == StateFailed || state == StateUp ||
		state == StateDown
}
