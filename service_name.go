package userconfig

import (
	"encoding/json"
	"regexp"
	"strings"
)

var (
	serviceNameRegExp = regexp.MustCompile("^[a-zA-Z0-9]{1}[a-z0-9A-Z_.-]{0,99}$")
)

type ServiceName string

// ParseServiceName returns the name of the given definition if it exists. It
// is does not exist, it generates an service name.
func ParseServiceName(b []byte) (string, error) {
	// try to fetch the name from a simple definition
	var simple map[string]interface{}
	if err := json.Unmarshal(b, &simple); err != nil {
		return "", mask(err)
	}

	for k, v := range simple {
		if strings.EqualFold(k, "name") {
			if name, ok := v.(string); ok {
				return name, nil
			} else {
				// name is not a string, break and return original error
				break
			}
		}
	}

	serviceDef, err := ParseServiceDefinition(b)
	if err != nil {
		return "", mask(err)
	}

	name, err := serviceDef.Name()
	if err != nil {
		return "", mask(err)
	}

	return name, nil
}

// String returns a string version of the given ServiceName.
func (an ServiceName) String() string {
	return string(an)
}

// Empty returns true if the given ServiceName is empty, false otherwise.
func (an ServiceName) Empty() bool {
	return an == ""
}

// Equals returns true if the given ServiceName is equal to the other
// given service name, false otherwise.
func (an ServiceName) Equals(other ServiceName) bool {
	return an == other
}

// Validate checks that the given ServiceName is a valid ServiceName.
func (an ServiceName) Validate() error {
	if an.Empty() {
		return maskf(InvalidServiceNameError, "service name must not be empty")
	}

	anStr := an.String()
	if !serviceNameRegExp.MatchString(anStr) {
		return maskf(InvalidServiceNameError, "service name '%s' must match regexp: %s", anStr, serviceNameRegExp)
	}

	return nil
}
