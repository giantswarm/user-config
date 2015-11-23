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

// TestLinkDetectCycleIssueWrtService: This test ensures that the link cycle detection
// works well with links to a component that link to another service.
func TestLinkDetectCycleIssueWrtService(t *testing.T) {
	// In this service, component 'a' links to component 'b' and component 'b'
	// links to another service. This should NOT result in validation errors.
	app := V2AppDefinition{
		Components: ComponentDefinitions{
			ComponentName("a"): &ComponentDefinition{
				Image: MustParseImageDefinition("registry.giantswarm.io/landingpage:0.10.0"),
				Links: LinkDefinitions{
					LinkDefinition{
						Component:  ComponentName("b"),
						TargetPort: generictypes.MustParseDockerPort("80/tcp"),
					},
				},
			},
			ComponentName("b"): &ComponentDefinition{
				Image: MustParseImageDefinition("registry.giantswarm.io/landingpage:0.10.0"),
				Links: LinkDefinitions{
					LinkDefinition{
						Service:    AppName("other-service"),
						TargetPort: generictypes.MustParseDockerPort("80/tcp"),
					},
				},
				Ports: PortDefinitions{
					generictypes.MustParseDockerPort("80/tcp"),
				},
			},
		},
	}
	valCtx := &ValidationContext{
		Protocols:     []string{"tcp"},
		MinVolumeSize: "1 GB",
		MaxVolumeSize: "100 GB",
		MinScaleSize:  1,
		MaxScaleSize:  10,
	}
	if err := app.Validate(valCtx); err != nil {
		t.Fatalf("Validate failed: %#v", err)
	}
}
