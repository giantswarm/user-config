package userconfig_test

import (
	"strings"
	"testing"

	"github.com/giantswarm/generic-types-go"
	"github.com/giantswarm/user-config"
)

func Test_CyclicDeps_Valid_ComponentLinks(t *testing.T) {
	def := userconfig.V2AppDefinition{
		Components: userconfig.ComponentDefinitions{},
	}

	// Component "one" links to component "two"
	def.Components[userconfig.ComponentName("one")] = &userconfig.ComponentDefinition{
		Image: userconfig.MustParseImageDefinition("registry.giantswarm.io/landingpage:0.10.0"),
		Ports: []generictypes.DockerPort{
			generictypes.MustParseDockerPort("80/tcp"),
		},
		Links: userconfig.LinkDefinitions{
			userconfig.LinkDefinition{
				Component:  userconfig.ComponentName("two"),
				TargetPort: generictypes.MustParseDockerPort("80/tcp"),
			},
		},
	}

	// Component "two" links to component "three"
	def.Components[userconfig.ComponentName("two")] = &userconfig.ComponentDefinition{
		Image: userconfig.MustParseImageDefinition("registry.giantswarm.io/landingpage:0.10.0"),
		Ports: []generictypes.DockerPort{
			generictypes.MustParseDockerPort("80/tcp"),
		},
		Links: userconfig.LinkDefinitions{
			userconfig.LinkDefinition{
				Component:  userconfig.ComponentName("three"),
				TargetPort: generictypes.MustParseDockerPort("80/tcp"),
			},
		},
	}

	// Component "three" links NOT to component "one"
	def.Components[userconfig.ComponentName("three")] = &userconfig.ComponentDefinition{
		Image: userconfig.MustParseImageDefinition("registry.giantswarm.io/landingpage:0.10.0"),
		Ports: []generictypes.DockerPort{
			generictypes.MustParseDockerPort("80/tcp"),
		},
	}

	if err := def.Validate(nil); err != nil {
		t.Fatalf("expected definition to be valid, got error: %#v", err)
	}
}

func Test_CyclicDeps_Valid_ServiceLinks(t *testing.T) {
	def := userconfig.V2AppDefinition{
		Components: userconfig.ComponentDefinitions{},
	}

	// Component "one" links to service "service"
	def.Components[userconfig.ComponentName("one")] = &userconfig.ComponentDefinition{
		Image: userconfig.MustParseImageDefinition("registry.giantswarm.io/landingpage:0.10.0"),
		Ports: []generictypes.DockerPort{
			generictypes.MustParseDockerPort("80/tcp"),
		},
		Links: userconfig.LinkDefinitions{
			userconfig.LinkDefinition{
				Service:    userconfig.AppName("service"),
				TargetPort: generictypes.MustParseDockerPort("80/tcp"),
			},
		},
	}

	// Component "two" links to component "three"
	def.Components[userconfig.ComponentName("two")] = &userconfig.ComponentDefinition{
		Image: userconfig.MustParseImageDefinition("registry.giantswarm.io/landingpage:0.10.0"),
		Ports: []generictypes.DockerPort{
			generictypes.MustParseDockerPort("80/tcp"),
		},
		Links: userconfig.LinkDefinitions{
			userconfig.LinkDefinition{
				Component:  userconfig.ComponentName("three"),
				TargetPort: generictypes.MustParseDockerPort("80/tcp"),
			},
		},
	}

	// Component "three" links NOT to component "one"
	def.Components[userconfig.ComponentName("three")] = &userconfig.ComponentDefinition{
		Image: userconfig.MustParseImageDefinition("registry.giantswarm.io/landingpage:0.10.0"),
		Ports: []generictypes.DockerPort{
			generictypes.MustParseDockerPort("80/tcp"),
		},
	}

	if err := def.Validate(nil); err != nil {
		t.Fatalf("expected definition to be valid, got error: %#v", err)
	}
}

func Test_CyclicDeps_LinkToSelf(t *testing.T) {
	def := userconfig.V2AppDefinition{
		Components: userconfig.ComponentDefinitions{},
	}

	// Component "one" links to component "two"
	def.Components[userconfig.ComponentName("one")] = &userconfig.ComponentDefinition{
		Image: userconfig.MustParseImageDefinition("registry.giantswarm.io/landingpage:0.10.0"),
		Ports: []generictypes.DockerPort{
			generictypes.MustParseDockerPort("80/tcp"),
		},
		Links: userconfig.LinkDefinitions{
			userconfig.LinkDefinition{
				Component:  userconfig.ComponentName("one"),
				TargetPort: generictypes.MustParseDockerPort("80/tcp"),
			},
		},
	}

	if err := def.Validate(nil); !userconfig.IsInvalidComponentDefinition(err) {
		t.Fatalf("expected error to be InvalidComponentDefinitionError, got: %#v", err)
	} else if !strings.Contains(err.Error(), userconfig.LinkCycleError.Error()) {
		t.Fatalf("expected error to provide message of LinkCycleError, got: %s", err.Error())
	}
}

func Test_CyclicDeps_OnlyCircle(t *testing.T) {
	def := userconfig.V2AppDefinition{
		Components: userconfig.ComponentDefinitions{},
	}

	// Component "one" links to itself
	def.Components[userconfig.ComponentName("one")] = &userconfig.ComponentDefinition{
		Image: userconfig.MustParseImageDefinition("registry.giantswarm.io/landingpage:0.10.0"),
		Ports: []generictypes.DockerPort{
			generictypes.MustParseDockerPort("80/tcp"),
		},
		Links: userconfig.LinkDefinitions{
			userconfig.LinkDefinition{
				Component:  userconfig.ComponentName("one"),
				TargetPort: generictypes.MustParseDockerPort("80/tcp"),
			},
		},
	}

	if err := def.Validate(nil); err == nil {
		t.Fatalf("expected cyclic dependencies to be detected and throw error")
	}

	// Component "one" links to component "two"
	def.Components[userconfig.ComponentName("one")] = &userconfig.ComponentDefinition{
		Image: userconfig.MustParseImageDefinition("registry.giantswarm.io/landingpage:0.10.0"),
		Ports: []generictypes.DockerPort{
			generictypes.MustParseDockerPort("80/tcp"),
		},
		Links: userconfig.LinkDefinitions{
			userconfig.LinkDefinition{
				Component:  userconfig.ComponentName("two"),
				TargetPort: generictypes.MustParseDockerPort("80/tcp"),
			},
		},
	}

	// Component "two" links to component "three"
	def.Components[userconfig.ComponentName("two")] = &userconfig.ComponentDefinition{
		Image: userconfig.MustParseImageDefinition("registry.giantswarm.io/landingpage:0.10.0"),
		Ports: []generictypes.DockerPort{
			generictypes.MustParseDockerPort("80/tcp"),
		},
		Links: userconfig.LinkDefinitions{
			userconfig.LinkDefinition{
				Component:  userconfig.ComponentName("three"),
				TargetPort: generictypes.MustParseDockerPort("80/tcp"),
			},
		},
	}

	// Component "three" links to component "one"
	def.Components[userconfig.ComponentName("three")] = &userconfig.ComponentDefinition{
		Image: userconfig.MustParseImageDefinition("registry.giantswarm.io/landingpage:0.10.0"),
		Ports: []generictypes.DockerPort{
			generictypes.MustParseDockerPort("80/tcp"),
		},
		Links: userconfig.LinkDefinitions{
			userconfig.LinkDefinition{
				Component:  userconfig.ComponentName("one"),
				TargetPort: generictypes.MustParseDockerPort("80/tcp"),
			},
		},
	}

	if err := def.Validate(nil); !userconfig.IsInvalidComponentDefinition(err) {
		t.Fatalf("expected error to be InvalidComponentDefinitionError, got: %#v", err)
	} else if !strings.Contains(err.Error(), userconfig.LinkCycleError.Error()) {
		t.Fatalf("expected error to provide message of LinkCycleError, got: %s", err.Error())
	}
}

func Test_CyclicDeps_AdditionalCircle(t *testing.T) {
	def := V2ExampleDefinition()

	// Component "one" links to component "two"
	def.Components[userconfig.ComponentName("one")] = &userconfig.ComponentDefinition{
		Image: userconfig.MustParseImageDefinition("registry.giantswarm.io/landingpage:0.10.0"),
		Ports: []generictypes.DockerPort{
			generictypes.MustParseDockerPort("80/tcp"),
		},
		Links: userconfig.LinkDefinitions{
			userconfig.LinkDefinition{
				Component:  userconfig.ComponentName("two"),
				TargetPort: generictypes.MustParseDockerPort("80/tcp"),
			},
		},
	}

	// Component "two" links to component "three"
	def.Components[userconfig.ComponentName("two")] = &userconfig.ComponentDefinition{
		Image: userconfig.MustParseImageDefinition("registry.giantswarm.io/landingpage:0.10.0"),
		Ports: []generictypes.DockerPort{
			generictypes.MustParseDockerPort("80/tcp"),
		},
		Links: userconfig.LinkDefinitions{
			userconfig.LinkDefinition{
				Component:  userconfig.ComponentName("three"),
				TargetPort: generictypes.MustParseDockerPort("80/tcp"),
			},
		},
	}

	// Component "three" links to component "one"
	def.Components[userconfig.ComponentName("three")] = &userconfig.ComponentDefinition{
		Image: userconfig.MustParseImageDefinition("registry.giantswarm.io/landingpage:0.10.0"),
		Ports: []generictypes.DockerPort{
			generictypes.MustParseDockerPort("80/tcp"),
		},
		Links: userconfig.LinkDefinitions{
			userconfig.LinkDefinition{
				Component:  userconfig.ComponentName("one"),
				TargetPort: generictypes.MustParseDockerPort("80/tcp"),
			},
		},
	}

	if err := def.Validate(nil); !userconfig.IsInvalidComponentDefinition(err) {
		t.Fatalf("expected error to be InvalidComponentDefinitionError, got: %#v", err)
	} else if !strings.Contains(err.Error(), userconfig.LinkCycleError.Error()) {
		t.Fatalf("expected error to provide message of LinkCycleError, got: %s", err.Error())
	}
}
