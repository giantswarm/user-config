package userconfig_test

import (
	"reflect"
	"strings"
	"testing"

	"github.com/giantswarm/generic-types-go"
	"github.com/giantswarm/user-config"
)

func Test_CyclicDeps_Valid_ComponentLinks(t *testing.T) {
	def := userconfig.ServiceDefinition{
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
	def := userconfig.ServiceDefinition{
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
				Service:    userconfig.ServiceName("service"),
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
	def := userconfig.ServiceDefinition{
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

	if err := def.Validate(nil); !userconfig.IsInvalidComponentDefinition(err) {
		t.Fatalf("expected error to be InvalidComponentDefinitionError, got: %#v", err)
	} else if !strings.Contains(err.Error(), userconfig.LinkCycleError.Error()) {
		t.Fatalf("expected error to provide message of LinkCycleError, got: %s", err.Error())
	}
}

func Test_CyclicDeps_OnlyCircle(t *testing.T) {
	def := userconfig.ServiceDefinition{
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
	def := ExampleDefinition()

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

func Test_AllDefsPerPod(t *testing.T) {
	service := testService()

	service.Components["root"] = setPod(testComponent(), userconfig.PodChildren)
	service.Components["root/a"] = testComponent()
	service.Components["root/b"] = testComponent()

	service.Components["other"] = setPod(testComponent(), userconfig.PodChildren)
	service.Components["other/a"] = testComponent()
	service.Components["other/b"] = testComponent()

	service.Components["a"] = testComponent()
	service.Components["a/b"] = testComponent()
	service.Components["a/c"] = testComponent()

	expectedDefsPerPod := []userconfig.ComponentDefinitions{
		userconfig.ComponentDefinitions{
			"root/a": service.Components["root/a"],
			"root/b": service.Components["root/b"],
		},
	}

	input := userconfig.ComponentDefinitions{
		"root/b": service.Components["root/b"],
	}

	defsPerPod, err := service.Components.AllDefsPerPod(input.ComponentNames())
	if err != nil {
		t.Fatalf("AllDefsPerPod failed: %#v", err)
	}
	if !reflect.DeepEqual(defsPerPod, expectedDefsPerPod) {
		t.Logf("defsPerPod:         %#v", defsPerPod)
		t.Logf("expectedDefsPerPod: %#v", expectedDefsPerPod)
		t.Fatalf("AllDefsPerPod failed: unexpected result")
	}
}
