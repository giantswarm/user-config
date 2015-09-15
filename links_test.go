package userconfig_test

import (
	"testing"

	"github.com/giantswarm/generic-types-go"
	. "github.com/giantswarm/user-config"
)

// TestLinkStringReliability ensures that LinkDefinitions.String always returns
// the same string. We use this to create proper diffs between two definitions,
// so this is quiet critical.
func TestLinkStringReliability(t *testing.T) {
	lds := LinkDefinitions{
		LinkDefinition{
			Service:    AppName("service1"),
			Component:  ComponentName("component1"),
			Alias:      "alias1",
			TargetPort: generictypes.MustParseDockerPort("80/tcp"),
		},
		LinkDefinition{
			Service:    AppName("aaa"),
			Component:  ComponentName("zzz"),
			Alias:      "bbb",
			TargetPort: generictypes.MustParseDockerPort("1111/tcp"),
		},
		LinkDefinition{
			Service:    AppName("foo"),
			Component:  ComponentName("bar"),
			Alias:      "baz",
			TargetPort: generictypes.MustParseDockerPort("9999/tcp"),
		},
	}

	expected := lds.String()

	for i := 0; i < 1000; i++ {
		generated := lds.String()

		if expected != generated {
			t.Log("expected link definitions to be qual")
			t.Logf("epxected: %s", expected)
			t.Fatalf("got: %s", generated)
		}
	}
}
