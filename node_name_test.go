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
