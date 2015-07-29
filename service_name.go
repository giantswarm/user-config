package userconfig

import (
	"regexp"
)

var (
	serviceNameRegExp = regexp.MustCompile("^[a-zA-Z0-9]{1}[a-z0-9A-Z_-]{0,99}$")
)

type ServiceName string

// String returns a string version of the given ServiceName.
func (sn ServiceName) String() string {
	return string(sn)
}

// Empty returns true if the given ServiceName is empty, false otherwise.
func (sn ServiceName) Empty() bool {
	return sn == ""
}

// Equals returns true if the given AppName is equal to the other
// given application name, false otherwise.
func (sn ServiceName) Equals(other ServiceName) bool {
	return sn == other
}

// Validate checks that the given ServiceName is a valid ServiceName.
func (sn ServiceName) Validate() error {
	if sn.Empty() {
		return maskf(InvalidServiceNameError, "service name must not be empty")
	}

	anStr := sn.String()
	if !serviceNameRegExp.MatchString(anStr) {
		return maskf(InvalidServiceNameError, "service name '%s' must match regexp: %s", anStr, serviceNameRegExp)
	}

	return nil
}
