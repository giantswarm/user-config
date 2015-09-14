package userconfig_test

import (
	"testing"

	"github.com/giantswarm/generic-types-go"
	. "github.com/giantswarm/user-config"
)

func TestExpose(t *testing.T) {
	list := []struct {
		Expose   ExposeDefinition
		ImplName string
		ImplPort string
	}{
		// Empty component, empty component port should refer to self with identical port
		{Expose: ExposeDefinition{Port: generictypes.MustParseDockerPort("80/tcp")}, ImplName: "self", ImplPort: "80/tcp"},
		// Set component, empty component port should refer to component with identical port
		{Expose: ExposeDefinition{Port: generictypes.MustParseDockerPort("80/tcp"), Component: ComponentName("foo")}, ImplName: "foo", ImplPort: "80/tcp"},
		// Empty component, specified component port should refer to self with specified port
		{Expose: ExposeDefinition{Port: generictypes.MustParseDockerPort("80/tcp"), TargetPort: generictypes.MustParseDockerPort("8080/tcp")}, ImplName: "self", ImplPort: "8080/tcp"},
	}

	for _, test := range list {
		implName := test.Expose.ImplementationComponentName(ComponentName("self"))
		if !implName.Equals(ComponentName(test.ImplName)) {
			t.Fatalf("invalid impl name detected: got '%s', expected '%s'", implName, test.ImplName)
		}

		implPort := test.Expose.ImplementationPort()
		if !implPort.Equals(generictypes.MustParseDockerPort(test.ImplPort)) {
			t.Fatalf("invalid impl port detected: got '%s', expected '%s'", implPort.String(), test.ImplPort)
		}
	}
}

func TestExposeResolveRecursive(t *testing.T) {
	implPorts := []generictypes.DockerPort{
		generictypes.MustParseDockerPort("80/tcp"),   // Same as exposed port
		generictypes.MustParseDockerPort("8086/tcp"), // Different from exposed port
	}
	// Test resolve of an ExposeDefinition that is implemented by the component itself.
	for _, implPort := range implPorts {
		nds := ComponentDefinitions{}
		nds["self"] = &ComponentDefinition{
			Image: MustParseImageDefinition("busybox"),
			Ports: PortDefinitions{
				implPort,
			},
			Expose: ExposeDefinitions{
				ExposeDefinition{
					Port:       generictypes.MustParseDockerPort("80/tcp"),
					TargetPort: implPort,
				},
			},
		}

		self, err := nds.ComponentByName("self")
		if err != nil {
			t.Fatalf("Cannot get self %#v", err)
		}

		resolvedName, resolvedPort, err := self.Expose[0].Resolve("self", nds)
		if err != nil {
			t.Fatalf("Cannot resolve self.Expose %#v", err)
		}

		if !resolvedName.Equals("self") {
			t.Fatal("Resolve yielded wrong implementation component")
		}

		if !resolvedPort.Equals(implPort) {
			t.Fatal("Resolve yielded wrong implementation port")
		}
	}
}

func TestExposeResolve(t *testing.T) {
	// Test that verifies that ExposeDefinition.Resolve uses absolute names
	// and handles recursion correctly.
	nds := ComponentDefinitions{}
	nds["a"] = &ComponentDefinition{
		Expose: ExposeDefinitions{
			ExposeDefinition{
				Port:       generictypes.MustParseDockerPort("80/tcp"),
				TargetPort: generictypes.MustParseDockerPort("8086/tcp"),
				Component:  "a/a",
			},
		},
	}
	nds["a/a"] = &ComponentDefinition{
		Expose: ExposeDefinitions{
			ExposeDefinition{
				Port:       generictypes.MustParseDockerPort("8086/tcp"),
				TargetPort: generictypes.MustParseDockerPort("85/tcp"),
				Component:  "a/a/a",
			},
		},
		Image: MustParseImageDefinition("busybox"),
		Ports: PortDefinitions{
			generictypes.MustParseDockerPort("8086/tcp"),
		},
	}
	nds["a/a/a"] = &ComponentDefinition{
		Image: MustParseImageDefinition("busybox"),
		Ports: PortDefinitions{
			generictypes.MustParseDockerPort("85/tcp"),
		},
	}

	a, err := nds.ComponentByName("a")
	if err != nil {
		t.Fatalf("Cannot get a %#v", err)
	}

	resolvedName, resolvedPort, err := a.Expose[0].Resolve("a", nds)
	if err != nil {
		t.Fatalf("Cannot resolve a.Expose %#v", err)
	}

	if !resolvedName.Equals("a/a/a") {
		t.Fatal("Resolve yielded wrong implementation component")
	}

	if !resolvedPort.Equals(generictypes.MustParseDockerPort("85/tcp")) {
		t.Fatal("Resolve yielded wrong implementation port")
	}
}

// TestExposeStringReliability ensures that ExposeDefinitions.String always returns
// the same string. We use this to create proper diffs between two definitions,
// so this is quiet critical.
func TestExposeStringReliability(t *testing.T) {
	eds := ExposeDefinitions{
		ExposeDefinition{
			Port:       generictypes.MustParseDockerPort("90/tcp"),
			Component:  ComponentName("component1"),
			TargetPort: generictypes.MustParseDockerPort("80/tcp"),
		},
		ExposeDefinition{
			Port:       generictypes.MustParseDockerPort("1111/tcp"),
			Component:  ComponentName("zzz"),
			TargetPort: generictypes.MustParseDockerPort("1111/tcp"),
		},
		ExposeDefinition{
			Port:       generictypes.MustParseDockerPort("8888/tcp"),
			Component:  ComponentName("bar"),
			TargetPort: generictypes.MustParseDockerPort("9999/tcp"),
		},
	}

	expected := eds.String()

	for i := 0; i < 1000; i++ {
		generated := eds.String()

		if expected != generated {
			t.Log("expected expose definitions to be qual")
			t.Logf("epxected: %s", expected)
			t.Fatalf("got: %s", generated)
		}
	}
}
