// TODO code within this file is deprecated. See
// https://github.com/giantswarm/container-states/

package userconfig

type Status string

const (
	STATUS_FAILED   Status = "failed"
	STATUS_DOWN     Status = "down"
	STATUS_STARTING Status = "starting"
	STATUS_UP       Status = "up"
	STATUS_STOPPING Status = "stopping"
)

func (s *Status) String() string {
	return string(*s)
}

var (
	// StatusOrder is used by AggregateStatus to determine the higher status.
	// Lower Index wins over higher Index
	StatusOrder = []Status{
		STATUS_FAILED,
		STATUS_DOWN,
		STATUS_STARTING,
		STATUS_STOPPING,
		STATUS_UP,
	}
)

// AggregateStatus returns the 'higher' of the two status, given the following order:
//  ok < stopping < starting < down < failed
func AggregateStatus(status1, status2 Status) Status {
	for _, nextStatus := range StatusOrder {
		if status1 == nextStatus || status2 == nextStatus {
			return nextStatus
		}
	}

	panic("unknown status: " + status1 + ", " + status2)
}

// Inactive means an app is failed or down.
func IsStatusInactive(status Status) bool {
	return status == STATUS_FAILED || status == STATUS_DOWN
}

// IsStatusFinal returns whether the given status is a final status and should not change upon itself.
// E.g. A unit with a status STARTING will after some time either switch to UP or FAILED,
// thus the state is not final.
// Final states are UP, DOWN or FAILED.
func IsStatusFinal(status Status) bool {
	return status == STATUS_FAILED || status == STATUS_UP || status == STATUS_DOWN
}
