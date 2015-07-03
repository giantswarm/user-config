package userconfig

import (
	"regexp"
	"strings"
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

// ParentName returns the parent name of the given name, or InvalidArgumentError if the name has no parent.
func (nn NodeName) ParentName() (NodeName, error) {
	parts := strings.Split(nn.String(), "/")
	if len(parts) > 1 {
		parts = parts[:len(parts)-1]
		parentName := strings.Join(parts, "/")
		return NodeName(parentName), nil
	}
	return NodeName(""), maskf(InvalidArgumentError, "'%s' has no parent", nn.String())
}
