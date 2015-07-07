package userconfig

import (
	"regexp"
	"strings"
)

var (
	nodeNameRegExp   = regexp.MustCompile("^[a-zA-Z0-9]{1}[a-z0-9A-Z_/-]{0,99}$")
	lettersAndDigits = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
)

type NodeName string

// String returns a string version of the given NodeName.
func (nn NodeName) String() string {
	return string(nn)
}

// Empty returns true if the given NodeName is empty, false otherwise.
func (nn NodeName) Empty() bool {
	return string(nn) == ""
}

// Validate checks that the given NodeName is a valid NodeName.
func (nn NodeName) Validate() error {
	if nn.Empty() {
		return maskf(InvalidNodeNameError, "node name must not be empty")
	}

	nnStr := nn.String()
	if !nodeNameRegExp.MatchString(nnStr) {
		return maskf(InvalidNodeNameError, "node name '%s' must match regexp: %s", nnStr, nodeNameRegExp)
	}

	if strings.HasSuffix(nnStr, "/") {
		return maskf(InvalidNodeNameError, "node name '%s' must not end with '/'", nnStr)
	}

	parts := strings.Split(nn.String(), "/")
	for _, part := range parts {
		if part == "" {
			return maskf(InvalidNodeNameError, "node name '%s' must not have empty parts", nnStr)
		}
		if !strings.ContainsAny(part, lettersAndDigits) {
			return maskf(InvalidNodeNameError, "node name '%s' (part '%s') must contain at least one letter or digit", nnStr, part)
		}
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

// IsDirectChildOf returns true if the given child name is a direct child of the given parent name.
// E.g.
// - "a/b".IsDirectChildOf("a") -> true
// - "a/b/c".IsDirectChildOf("a") -> false
func (childName NodeName) IsDirectChildOf(parentName NodeName) bool {
	prefix := parentName.String()
	if !strings.HasSuffix(prefix, "/") {
		prefix = prefix + "/"
	}
	childNameStr := childName.String()
	if !strings.HasPrefix(childNameStr, prefix) {
		return false
	}
	name := childNameStr[len(prefix):]
	if strings.Contains(name, "/") {
		// Grand child
		return false
	}
	return true
}

// IsChildOf returns true if the given child name is a child (recursive) of the given parent name.
// E.g.
// - "a/b".IsChildOf("a") -> true
// - "a/b/c".IsChildOf("a") -> true
func (childName NodeName) IsChildOf(parentName NodeName) bool {
	prefix := parentName.String()
	if !strings.HasSuffix(prefix, "/") {
		prefix = prefix + "/"
	}
	childNameStr := childName.String()
	if !strings.HasPrefix(childNameStr, prefix) {
		return false
	}
	return true
}

// IsSiblingOf returns true if the given other name is a sibling of the given name.
// E.g.
// - "a/b".IsSiblingOf("a") -> false
// - "a/c".IsSiblingOf("a/b") -> true
// - "a/b/c".IsSiblingOf("a/b") -> false
func (name NodeName) IsSiblingOf(otherName NodeName) bool {
	parentName, err := name.ParentName()
	if err != nil {
		parentName = ""
	}
	otherParentName, err := otherName.ParentName()
	if err != nil {
		otherParentName = ""
	}

	return parentName == otherParentName
}
