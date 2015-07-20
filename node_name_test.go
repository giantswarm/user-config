package userconfig_test

import (
	"testing"

	"github.com/giantswarm/user-config"
)

func TestValidNodeNames(t *testing.T) {
	list := map[string]string{
		"a":                   "node name should be allowed to be normal single character",
		"x":                   "node name should be allowed to be normal single character",
		"0":                   "node name should be allowed to be normal single character",
		"3":                   "node name should be allowed to be normal single character",
		"wjehfg":              "node name should be allowed to contain normal words",
		"a/b/c":               "node name should be allowed to be path",
		"0/1/2":               "node name should be allowed to be path",
		"wjehfg/skdjcsd/jshg": "node name should be allowed to be path",
		"a-0/b-1/c-2":         "node name should be allowed to be path containing special chars",
	}

	for name, reason := range list {
		nodeName := userconfig.NodeName(name)
		err := nodeName.Validate()

		if err != nil {
			t.Fatalf("valid node name '%s' detected to be invalid: %s", name, reason)
		}
	}
}

func TestInvalidNodeNames(t *testing.T) {
	list := map[string]string{
		"":      "node name must not be empty",
		"-":     "node name must not start with special chars",
		"-/-/-": "node name must not start with special chars",
		"_/-":   "node name must not start with special chars",
		"/_/-":  "node name must not start with special chars",
		"///":   "node name must not start with special chars",
		"/a":    "node name must not start with special chars",
		"-x":    "node name must not start with special chars",
		"&0":    "node name must not start with special chars",
		"$3":    "node name must not start with special chars",
		"()wjehfg/skdjcsd/jshg": "node name must not start with special chars",
		"-a-0/b-1/c-2":          "node name must not start with special chars",
		"a/b/c/":                "node name must not end with '/'",
		"/a/b/c":                "node name must not start with '/'",
		"a//b/c":                "node name must not contain '//'",
		"a/---/b/c":             "node name parts must contain at least one letter or digit",
		"/":                     "node name must not start with '/'",
		" ":                     "node name parts must contain at least one letter or digit",
		"a ":                    "node name parts must not contain spaces",
	}

	for name, reason := range list {
		nodeName := userconfig.NodeName(name)
		err := nodeName.Validate()

		if err == nil {
			t.Fatalf("invalid node name '%s' not detected: %s", name, reason)
		}
		if !userconfig.IsInvalidNodeName(err) {
			t.Fatalf("expected error to be InvalidNodeNameError")
		}
	}
}

func TestNodeNameParentName(t *testing.T) {
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
		child := userconfig.NodeName(test.Name)
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

func TestNodeNameLocalName(t *testing.T) {
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
		child := userconfig.NodeName(test.Name)
		local := child.LocalName()
		if local.String() != test.LocalName {
			t.Fatalf("Test %v failed: got '%s', expected '%s'", test, local.String(), test.LocalName)
		}
	}
}

func TestNodeNameIsDirectChild(t *testing.T) {
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
		child := userconfig.NodeName(test.ChildName)
		parent := userconfig.NodeName(test.ParentName)
		result := child.IsDirectChildOf(parent)
		if test.Result != result {
			t.Fatalf("Test %v failed: got '%v', expected '%v'", test, result, test.Result)
		}
	}
}

func TestNodeNameIsChild(t *testing.T) {
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
		child := userconfig.NodeName(test.ChildName)
		parent := userconfig.NodeName(test.ParentName)
		result := child.IsChildOf(parent)
		if test.Result != result {
			t.Fatalf("Test %v failed: got '%v', expected '%v'", test, result, test.Result)
		}
	}
}

func TestNodeNameIsSibling(t *testing.T) {
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
		name1 := userconfig.NodeName(test.Name1)
		name2 := userconfig.NodeName(test.Name2)
		result := name1.IsSiblingOf(name2)
		if test.Result != result {
			t.Fatalf("Test %v failed: got '%v', expected '%v'", test, result, test.Result)
		}
	}
}

func TestNodeNameEmpty(t *testing.T) {
	list := []struct {
		Name   string
		Result bool
	}{
		{"", true},
		{"a", false},
		{"a/b", false},
	}

	for _, test := range list {
		name := userconfig.NodeName(test.Name)
		result := name.Empty()
		if test.Result != result {
			t.Fatalf("Test %v failed: got '%v', expected '%v'", test, result, test.Result)
		}
	}
}
