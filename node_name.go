package userconfig

import (
	"regexp"
)

var (
	nodeNameRegExp = regexp.MustCompile("^[a-zA-Z0-9]{1}[a-z0-9A-Z_/-]{0,99}$")
)

type NodeName string

func (nn NodeName) String() string {
	return string(nn)
}

func (nn NodeName) Validate() error {
	if nn == "" {
		return maskf(InvalidNodeNameError, "node name must not be empty")
	}

	if !nodeNameRegExp.MatchString(nn.String()) {
		return maskf(InvalidNodeNameError, "node name '%s' must match regexp: %s", nn, nodeNameRegExp)
	}

	return nil
}
