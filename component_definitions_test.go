package userconfig_test

import (
	"testing"

	"github.com/giantswarm/generic-types-go"
	"github.com/giantswarm/user-config"
)

func Test_AllDefsPerPod_Sorting_NoPods(t *testing.T) {
	def := userconfig.ServiceDefinition{
		Components: userconfig.ComponentDefinitions{},
	}

	// Component "a" links to component "b", no pods involved
	// Therefore we expect to get 2 maps from AllDefsPerPod, the first containing 'b', the second containing 'a'
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

func Test_AllDefsPerPod_Sorting_WithPods(t *testing.T) {
	def := userconfig.ServiceDefinition{
		Components: userconfig.ComponentDefinitions{},
	}

	// Component "pod/a" links to component "pod/b"
	// Component "pod/b" links to component "nonpod"
	// Therefore we expect to get 2 maps from AllDefsPerPod, the first containint 'nonpod', the last containing 'pod/a' & 'pod/b'
	def.Components["nonpod"] = testComponent()
	def.Components["pod"] = testComponent()
	def.Components["pod/a"] = testComponent()
	def.Components["pod/b"] = testComponent()
	def.Components["pod"].Pod = "children"
	def.Components["nonpod"].Ports = userconfig.PortDefinitions{
		generictypes.MustParseDockerPort("88"),
	}
	def.Components["pod/b"].Ports = userconfig.PortDefinitions{
		generictypes.MustParseDockerPort("80"),
	}
	def.Components["pod/a"].Links = userconfig.LinkDefinitions{
		userconfig.LinkDefinition{
			Component:  userconfig.ComponentName("pod/b"),
			TargetPort: generictypes.MustParseDockerPort("80/tcp"),
		},
	}
	def.Components["pod/b"].Links = userconfig.LinkDefinitions{
		userconfig.LinkDefinition{
			Component:  userconfig.ComponentName("nonpod"),
			TargetPort: generictypes.MustParseDockerPort("88/tcp"),
		},
	}

	if err := def.Validate(nil); err != nil {
		t.Fatalf("expected definition to be valid, got error: %#v", err)
	}

	names := userconfig.ComponentNames{"pod/a", "pod/b", "nonpod"}
	defs, err := def.Components.AllDefsPerPod(names)
	if err != nil {
		t.Fatalf("AllDefsPerPod failed: %#v", err)
	}
	if len(defs) != 2 {
		t.Fatalf("Expected to get 2 maps, got %d", len(defs))
	}
	if !defs[0].Contains("nonpod") {
		t.Fatalf("Expected 'nonpod' to come first, got %#v", defs[0])
	}
	if !defs[1].Contains("pod/a") || !defs[1].Contains("pod/b") {
		t.Fatalf("Expected 'pod/a' & 'pod/b' to come last, got %#v", defs[1])
	}
}

func Test_AllDefsPerPod_Sorting_NotAllNames(t *testing.T) {
	def := userconfig.ServiceDefinition{
		Components: userconfig.ComponentDefinitions{},
	}

	// Component "box" links to component "dep", no pods involved, we only ask for "box"
	// Therefore we expect to get 1 map from AllDefsPerPod, containing 'box'.
	// No error is allowed.
	def.Components["dep"] = testComponent()
	def.Components["box"] = testComponent()
	def.Components["dep"].Ports = userconfig.PortDefinitions{
		generictypes.MustParseDockerPort("80"),
	}
	def.Components["box"].Links = userconfig.LinkDefinitions{
		userconfig.LinkDefinition{
			Component:  userconfig.ComponentName("dep"),
			TargetPort: generictypes.MustParseDockerPort("80/tcp"),
		},
	}

	if err := def.Validate(nil); err != nil {
		t.Fatalf("expected definition to be valid, got error: %#v", err)
	}

	names := userconfig.ComponentNames{"box"}
	defs, err := def.Components.AllDefsPerPod(names)
	if err != nil {
		t.Fatalf("AllDefsPerPod failed: %#v", err)
	}
	if len(defs) != 1 {
		t.Fatalf("Expected to get 1 maps, got %d", len(defs))
	}
	if !defs[0].Contains("box") {
		t.Fatalf("Expected 'box' to come first, got %#v", defs[0])
	}
}
