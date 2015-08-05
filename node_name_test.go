package userconfig_test

import (
	"testing"

	"github.com/giantswarm/user-config"
)

func TestValidComponentNames(t *testing.T) {
	list := map[string]string{
		"a":                   "component name should be allowed to be normal single character",
		"x":                   "component name should be allowed to be normal single character",
		"0":                   "component name should be allowed to be normal single character",
		"3":                   "component name should be allowed to be normal single character",
		"wjehfg":              "component name should be allowed to contain normal words",
		"a/b/c":               "component name should be allowed to be path",
		"0/1/2":               "component name should be allowed to be path",
		"wjehfg/skdjcsd/jshg": "component name should be allowed to be path",
		"a-0/b-1/c-2":         "component name should be allowed to be path containing special chars",
	}

	for name, reason := range list {
		componentName := userconfig.ComponentName(name)
		err := componentName.Validate()

		if err != nil {
			t.Fatalf("valid component name '%s' detected to be invalid: %s", name, reason)
		}
	}
}

func TestInvalidComponentNames(t *testing.T) {
	list := map[string]string{
		"":      "component name must not be empty",
		"-":     "component name must not start with special chars",
		"-/-/-": "component name must not start with special chars",
		"_/-":   "component name must not start with special chars",
		"/_/-":  "component name must not start with special chars",
		"///":   "component name must not start with special chars",
		"/a":    "component name must not start with special chars",
		"-x":    "component name must not start with special chars",
		"&0":    "component name must not start with special chars",
		"$3":    "component name must not start with special chars",
		"()wjehfg/skdjcsd/jshg": "component name must not start with special chars",
		"-a-0/b-1/c-2":          "component name must not start with special chars",
		"a/b/c/":                "component name must not end with '/'",
		"/a/b/c":                "component name must not start with '/'",
		"a//b/c":                "component name must not contain '//'",
		"a/---/b/c":             "component name parts must contain at least one letter or digit",
		"/":                     "component name must not start with '/'",
		" ":                     "component name parts must contain at least one letter or digit",
		"a ":                    "component name parts must not contain spaces",
	}

	for name, reason := range list {
		componentName := userconfig.ComponentName(name)
		err := componentName.Validate()

		if err == nil {
			t.Fatalf("invalid component name '%s' not detected: %s", name, reason)
		}
		if !userconfig.IsInvalidComponentName(err) {
			t.Fatalf("expected error to be InvalidComponentNameError")
		}
	}
}

func TestComponentNameParentName(t *testing.T) {
	list := []struct {
		Name       string
		ParentName string
		IsValid    bool
	}{
		{"a", "", false},
		{"a/b", "a", true},
		{"", "", false},
		{"a/b/c", "a/b", true},
	}

	for _, test := range list {
		child := userconfig.ComponentName(test.Name)
		parent, err := child.ParentName()
		if test.IsValid {
			if err != nil {
				t.Fatalf("Test %v failed %v", test, err)
			}
			if parent.String() != test.ParentName {
				t.Fatalf("Test %v failed: got '%s', expected '%s'", test, parent.String(), test.ParentName)
			}
		} else {
			if err == nil {
				t.Fatalf("Test %v succeeded while error expected", test)
			}
		}
	}
}

func TestComponentNameLocalName(t *testing.T) {
	list := []struct {
		Name      string
		LocalName string
	}{
		{"a", "a"},
		{"a/b", "b"},
		{"", ""},
		{"a/b/c", "c"},
	}

	for _, test := range list {
		child := userconfig.ComponentName(test.Name)
		local := child.LocalName()
		if local.String() != test.LocalName {
			t.Fatalf("Test %v failed: got '%s', expected '%s'", test, local.String(), test.LocalName)
		}
	}
}

func TestComponentNameIsDirectChild(t *testing.T) {
	list := []struct {
		ParentName string
		ChildName  string
		Result     bool
	}{
		{"a", "a/b", true},
		{"a/b", "a", false},
		{"a", "a/b/c", false},
		{"a", "b/c", false},
	}

	for _, test := range list {
		child := userconfig.ComponentName(test.ChildName)
		parent := userconfig.ComponentName(test.ParentName)
		result := child.IsDirectChildOf(parent)
		if test.Result != result {
			t.Fatalf("Test %v failed: got '%v', expected '%v'", test, result, test.Result)
		}
	}
}

func TestComponentNameIsChild(t *testing.T) {
	list := []struct {
		ParentName string
		ChildName  string
		Result     bool
	}{
		{"a", "a/b", true},
		{"a/b", "a", false},
		{"a", "a/b/c", true},
		{"a/b", "a/b/c/d/e/f", true},
		{"a", "b/c", false},
	}

	for _, test := range list {
		child := userconfig.ComponentName(test.ChildName)
		parent := userconfig.ComponentName(test.ParentName)
		result := child.IsChildOf(parent)
		if test.Result != result {
			t.Fatalf("Test %v failed: got '%v', expected '%v'", test, result, test.Result)
		}
	}
}

func TestComponentNameIsSibling(t *testing.T) {
	list := []struct {
		Name1  string
		Name2  string
		Result bool
	}{
		{"a", "a/b", false},
		{"a/b", "a", false},
		{"a/b", "a/c", true},
		{"a/b/c", "a/b/b", true},
		{"a/b/c", "a/c/b", false},
		{"a/b", "a", false},
		{"a", "c", true},
	}

	for _, test := range list {
		name1 := userconfig.ComponentName(test.Name1)
		name2 := userconfig.ComponentName(test.Name2)
		result := name1.IsSiblingOf(name2)
		if test.Result != result {
			t.Fatalf("Test %v failed: got '%v', expected '%v'", test, result, test.Result)
		}
	}
}

func TestComponentNameEmpty(t *testing.T) {
	list := []struct {
		Name   string
		Result bool
	}{
		{"", true},
		{"a", false},
		{"a/b", false},
	}

	for _, test := range list {
		name := userconfig.ComponentName(test.Name)
		result := name.Empty()
		if test.Result != result {
			t.Fatalf("Test %v failed: got '%v', expected '%v'", test, result, test.Result)
		}
	}
}
