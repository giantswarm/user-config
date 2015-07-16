package userconfig

import (
	"regexp"
)

var (
	appNameRegExp = regexp.MustCompile("^[a-zA-Z0-9]{1}[a-z0-9A-Z_-]{0,99}$")
)

type AppName string

// String returns a string version of the given AppName.
func (an AppName) String() string {
	return string(an)
}

// Empty returns true if the given AppName is empty, false otherwise.
func (an AppName) Empty() bool {
	return an == ""
}

// Equals returns true if the given AppName is equal to the other
// given application name, false otherwise.
func (an AppName) Equals(other AppName) bool {
	return an == other
}

// Validate checks that the given NodeName is a valid NodeName.
func (an AppName) Validate() error {
	if an.Empty() {
		return maskf(InvalidAppNameError, "application name must not be empty")
	}

	anStr := an.String()
	if !appNameRegExp.MatchString(anStr) {
		return maskf(InvalidAppNameError, "application name '%s' must match regexp: %s", anStr, appNameRegExp)
	}

	return nil
}
