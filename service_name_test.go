package userconfig_test

import (
	"testing"

	"github.com/giantswarm/user-config"
)

func TestValidServiceNames(t *testing.T) {
	list := map[string]string{
		"a":        "service name should be allowed to be normal single character",
		"x":        "service name should be allowed to be normal single character",
		"0":        "service name should be allowed to be normal single character",
		"3":        "service name should be allowed to be normal single character",
		"wjehfg":   "service name should be allowed to contain normal words",
		"wj_eh-fg": "service name should be allowed to contain dashes and underscores",
	}

	for name, reason := range list {
		serviceName := userconfig.ServiceName(name)
		err := serviceName.Validate()

		if err != nil {
			t.Fatalf("valid service name '%s' detected to be invalid: %s", name, reason)
		}
	}
}

func TestInvalidServiceNames(t *testing.T) {
	list := map[string]string{
		"":      "service name must not be empty",
		"-":     "service name must not start with special chars",
		"-/-/-": "service name must not start contain slashes",
		"-x":    "service name must not start with special chars",
		"&0":    "service name must not start with special chars",
		"$3":    "service name must not start with special chars",
		"()wjehfg/skdjcsd/jshg": "service name must not start with special chars",
		"a ": "service name parts must not contain spaces",
	}

	for name, reason := range list {
		serviceName := userconfig.ServiceName(name)
		err := serviceName.Validate()

		if err == nil {
			t.Fatalf("invalid service name '%s' not detected: %s", name, reason)
		}
		if !userconfig.IsInvalidServiceName(err) {
			t.Fatalf("expected error to be InvalidServiceNameError")
		}
	}
}

func TestServiceNameEmpty(t *testing.T) {
	list := []struct {
		Name   string
		Result bool
	}{
		{"", true},
		{"a", false},
		{"a/b", false},
	}

	for _, test := range list {
		name := userconfig.ServiceName(test.Name)
		result := name.Empty()
		if test.Result != result {
			t.Fatalf("Test %v failed: got '%v', expected '%v'", test, result, test.Result)
		}
	}
}
