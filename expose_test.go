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
