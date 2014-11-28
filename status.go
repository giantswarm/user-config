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

// AggregateStatus returns the 'higher' of the two status, given the following order:
//  ok < starting < stopping < down < failed
func AggregateStatus(status1, status2 Status) State {
	if status1.State == StateFailed || status2.State == StateFailed {
		return StateFailed
	}
	if status1.State == StateDown || status2.State == StateDown {
		return StateDown
	}
	if status1.State == StateStopping || status2.State == StateStopping {
		return StateStopping
	}
	if status1.State == StateStarting || status2.State == StateStarting {
		return StateStarting
	}
	return StateUp
}

// Active means starting or up.
func IsStatusActive(status Status) bool {
	return status.State == StateStarting || status.State == StateUp
}

// IsStatusFinal returns whether the given status is a final status and should
// not change upon itself. E.g. A unit with a StateStarting will after some
// time either switch to StateUp or StateFailed, thus the state is not final.
// Final states are StateUp, StateDown or StateFailed.
func IsStatusFinal(status Status) bool {
	return status.State == StateFailed || status.State == StateUp || status.State == StateDown
}
