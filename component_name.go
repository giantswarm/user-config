package userconfig

import (
	"regexp"
	"strings"
)

var (
	componentNameRegExp = regexp.MustCompile("^[a-zA-Z0-9]{1}[a-z0-9A-Z_/-]{0,99}$")
	lettersAndDigits    = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
)

type ComponentName string

// String returns a string version of the given ComponentName.
func (cn ComponentName) String() string {
	return string(cn)
}

// Empty returns true if the given ComponentName is empty, false otherwise.
func (cn ComponentName) Empty() bool {
	return cn == ""
}

// Equals returns true if the given ComponentName is equal to the other
// given component name, false otherwise.
func (cn ComponentName) Equals(other ComponentName) bool {
	return cn == other
}

// Validate checks that the given ComponentName is a valid ComponentName.
func (cn ComponentName) Validate() error {
	if cn.Empty() {
		return maskf(InvalidComponentNameError, "component name must not be empty")
	}

	nnStr := cn.String()
	if !componentNameRegExp.MatchString(nnStr) {
		return maskf(InvalidComponentNameError, "component name '%s' must match regexp: %s", nnStr, componentNameRegExp)
	}

	if strings.HasSuffix(nnStr, "/") {
		return maskf(InvalidComponentNameError, "component name '%s' must not end with '/'", nnStr)
	}

	parts := strings.Split(nnStr, "/")
	for _, part := range parts {
		if part == "" {
			return maskf(InvalidComponentNameError, "component name '%s' must not have empty parts", nnStr)
		}
		if !strings.ContainsAny(part, lettersAndDigits) {
			return maskf(InvalidComponentNameError, "component name '%s' (part '%s') must contain at least one letter or digit", nnStr, part)
		}
	}

	return nil
}

// ParentName returns the parent name of the given name, or InvalidArgumentError if the name has no parent.
func (cn ComponentName) ParentName() (ComponentName, error) {
	parts := strings.Split(cn.String(), "/")
	if len(parts) > 1 {
		parts = parts[:len(parts)-1]
		parentName := strings.Join(parts, "/")
		return ComponentName(parentName), nil
	}
	return ComponentName(""), maskf(InvalidArgumentError, "'%s' has no parent", cn.String())
}

// LocalName returns the last part of the given name.
func (cn ComponentName) LocalName() ComponentName {
	parts := strings.Split(cn.String(), "/")
	return ComponentName(parts[len(parts)-1])
}

// IsDirectChildOf returns true if the given child name is a direct child of the given parent name.
// E.g.
// - "a/b".IsDirectChildOf("a") -> true
// - "a/b/c".IsDirectChildOf("a") -> false
func (childName ComponentName) IsDirectChildOf(parentName ComponentName) bool {
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
func (childName ComponentName) IsChildOf(parentName ComponentName) bool {
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
func (name ComponentName) IsSiblingOf(otherName ComponentName) bool {
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
