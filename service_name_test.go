package userconfig_test

import (
	"testing"

	"github.com/giantswarm/user-config"
)

func TestValidServiceNames(t *testing.T) {
	list := map[string]string{
		"a":           "app name should be allowed to be normal single character",
		"x":           "app name should be allowed to be normal single character",
		"0":           "app name should be allowed to be normal single character",
		"3":           "app name should be allowed to be normal single character",
		"wjehfg":      "app name should be allowed to contain normal words",
		"wj_eh-fg":    "app name should be allowed to contain dashes and underscores",
		"foo-bar.com": "app name should be allowed to contain dashes and dots",
	}

	for name, reason := range list {
		appName := userconfig.ServiceName(name)
		err := appName.Validate()

		if err != nil {
			t.Fatalf("valid app name '%s' detected to be invalid: %s", name, reason)
		}
	}
}

func TestInvalidServiceNames(t *testing.T) {
	list := map[string]string{
		"":      "app name must not be empty",
		" ":     "app name must not be empty space",
		"-":     "app name must not start with special chars",
		".":     "app name must not start with special chars",
		"-/-/-": "app name must not start contain slashes",
		"-x":    "app name must not start with special chars",
		"&0":    "app name must not start with special chars",
		"$3":    "app name must not start with special chars",
		"()wjehfg/skdjcsd/jshg": "app name must not start with special chars",
		"a ":   "app name parts must not contain spaces",
		" a":   "app name parts must not start with spaces",
		".foo": "app name parts must not start with dots",
	}

	for name, reason := range list {
		appName := userconfig.ServiceName(name)
		err := appName.Validate()

		if err == nil {
			t.Fatalf("invalid app name '%s' not detected: %s", name, reason)
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
