package userconfig

import (
	"regexp"

	"github.com/juju/errgo"
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
		return mask(errgo.WithCausef(nil, InvalidNodeNameError, "node name must not be empty"))
	}

	if !nodeNameRegExp.MatchString(nn.String()) {
		return mask(errgo.WithCausef(nil, InvalidNodeNameError, "node name '%s' must match regexp: %s", nn, nodeNameRegExp))
	}

	return nil
}
