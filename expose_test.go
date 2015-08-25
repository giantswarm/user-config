package userconfig_test

import (
	"testing"

	"github.com/giantswarm/generic-types-go"
	"github.com/giantswarm/user-config"
)

func TestExpose(t *testing.T) {
	list := []struct {
		Expose   userconfig.ExposeDefinition
		ImplName string
		ImplPort string
	}{
		// Empty component, empty component port should refer to self with identical port
		{Expose: userconfig.ExposeDefinition{Port: generictypes.MustParseDockerPort("80/tcp")}, ImplName: "self", ImplPort: "80/tcp"},
		// Set component, empty component port should refer to component with identical port
		{Expose: userconfig.ExposeDefinition{Port: generictypes.MustParseDockerPort("80/tcp"), Component: userconfig.ComponentName("foo")}, ImplName: "foo", ImplPort: "80/tcp"},
		// Empty component, specified component port should refer to self with specified port
		{Expose: userconfig.ExposeDefinition{Port: generictypes.MustParseDockerPort("80/tcp"), TargetPort: generictypes.MustParseDockerPort("8080/tcp")}, ImplName: "self", ImplPort: "8080/tcp"},
	}

	for _, test := range list {
		implName := test.Expose.ImplementationComponentName(userconfig.ComponentName("self"))
		if !implName.Equals(userconfig.ComponentName(test.ImplName)) {
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
		nds := userconfig.ComponentDefinitions{}
		nds["self"] = &userconfig.ComponentDefinition{
			Image: userconfig.MustParseImageDefinition("busybox"),
			Ports: userconfig.PortDefinitions{
				implPort,
			},
			Expose: userconfig.ExposeDefinitions{
				userconfig.ExposeDefinition{
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
