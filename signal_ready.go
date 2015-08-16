package userconfig

const SignalReadyMechanism = "systemd-notify"

type SignalReady string

// String returns a string version of the given SignalReady.
func (sr SignalReady) String() string {
	return string(sr)
}

// Empty returns true if the given SignalReady is empty, false otherwise.
func (sr SignalReady) Empty() bool {
	return sr == ""
}

// Equals returns true if the given SignalReady is equal to the other
// given signal ready mechanism, false otherwise.
func (sr SignalReady) Equals(other SignalReady) bool {
	return sr == other
}

// Validate checks that the given SignalReady is a valid SignalReady.
func (sr SignalReady) Validate() error {
	if !sr.Empty() && !sr.Equals(SignalReadyMechanism) {
		return maskf(InvalidAppNameError, "signal ready must be 'systemd-notify'")
	}

	return nil
}
