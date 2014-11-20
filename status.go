package userconfig

type Status string

const (
	STATUS_FAILED   Status = "failed"
	STATUS_DOWN     Status = "down"
	STATUS_STARTING Status = "starting"
	STATUS_UP       Status = "up"
)

func (s *Status) String() string {
	return string(*s)
}

// AggregateStatus returns the 'higher' of the two status, given the following order:
//  ok < starting < down < failed
func AggregateStatus(status1, status2 Status) Status {
	if status1 == STATUS_FAILED || status2 == STATUS_FAILED {
		return STATUS_FAILED
	}
	if status1 == STATUS_DOWN || status2 == STATUS_DOWN {
		return STATUS_DOWN
	}
	if status1 == STATUS_STARTING || status2 == STATUS_STARTING {
		return STATUS_STARTING
	}
	return STATUS_UP
}

// Inactive means an app is failed or down.
func IsStatusInactive(status Status) bool {
	return status == STATUS_FAILED || status == STATUS_DOWN
}

// Active means starting or up.
func IsStatusActive(status Status) bool {
	return status == STATUS_STARTING || status == STATUS_UP
}

// Final means failed or up.
func IsStatusFinal(status Status) bool {
	return status == STATUS_FAILED || status == STATUS_UP
}
