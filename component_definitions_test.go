package userconfig_test

import (
	"testing"

	"github.com/giantswarm/generic-types-go"
	"github.com/giantswarm/user-config"
)

func Test_AllDefsPerPod_Sorting(t *testing.T) {
	def := userconfig.ServiceDefinition{
		Components: userconfig.ComponentDefinitions{},
	}

	// Component "one" links to component "two"
	def.Components["a"] = testComponent()
	def.Components["b"] = testComponent()
	def.Components["b"].Ports = userconfig.PortDefinitions{
		generictypes.MustParseDockerPort("80"),
	}
	def.Components["a"].Links = userconfig.LinkDefinitions{
		userconfig.LinkDefinition{
			Component:  userconfig.ComponentName("b"),
			TargetPort: generictypes.MustParseDockerPort("80/tcp"),
		},
	}

	if err := def.Validate(nil); err != nil {
		t.Fatalf("expected definition to be valid, got error: %#v", err)
	}

	names := userconfig.ComponentNames{"a", "b"}
	defs, err := def.Components.AllDefsPerPod(names)
	if err != nil {
		t.Fatalf("AllDefsPerPod failed: %#v", err)
	}
	if len(defs) != 2 {
		t.Fatalf("Expected to get 2 maps, got %d", len(defs))
	}
	if !defs[0].Contains("b") {
		t.Fatalf("Expected 'b' to come first, got %#v", defs[0])
	}
	if !defs[1].Contains("a") {
		t.Fatalf("Expected 'a' to come last, got %#v", defs[1])
	}
}
