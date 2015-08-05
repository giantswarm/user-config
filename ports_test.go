package userconfig_test

import (
	"encoding/json"
	"testing"

	"github.com/giantswarm/generic-types-go"
	"github.com/giantswarm/user-config"
)

func TestValidPortsValues(t *testing.T) {
	list := []struct {
		Input  string
		Result userconfig.PortDefinitions
	}{
		{`80`, userconfig.PortDefinitions{generictypes.MustParseDockerPort("80")}},
		{`"80/tcp"`, userconfig.PortDefinitions{generictypes.MustParseDockerPort("80")}},
		{`["80/tcp","81/tcp"]`, userconfig.PortDefinitions{generictypes.MustParseDockerPort("80"), generictypes.MustParseDockerPort("81")}},
		{`[80,"81/tcp","82"]`, userconfig.PortDefinitions{generictypes.MustParseDockerPort("80"), generictypes.MustParseDockerPort("81"), generictypes.MustParseDockerPort("82")}},
	}

	for _, test := range list {
		var pds userconfig.PortDefinitions
		if err := json.Unmarshal([]byte(test.Input), &pds); err != nil {
			t.Fatalf("Valid port definitions value '%s' considered invalid because %v", test.Input, err)
		}
		if len(pds) != len(test.Result) {
			t.Fatalf("Invalid length, expected %v, got %v", len(test.Result), len(pds))
		}
		for i, x := range pds {
			if !x.Equals(test.Result[i]) {
				t.Fatalf("Invalid element at %v, expected %v, got %v", i, test.Result[i], pds[i])
			}
		}
	}
}

func TestInvalidPortsValues(t *testing.T) {
	list := []string{
		``,
		`{"field":"foo"}`,
	}

	for _, s := range list {
		var pds userconfig.PortDefinitions
		if err := json.Unmarshal([]byte(s), &pds); err == nil {
			t.Fatalf("Invalid ports value '%s' considered valid", s)
		}
	}
}

func TestUnmarshalV2PortsFullService(t *testing.T) {
	// Test the validator for full service containing ports in various formats
	var appDef userconfig.V2AppDefinition

	byteSlice := []byte(`{
    "components": {
        "component1": {
            "image": "busybox",
            "ports": [ "80" ]
        },
        "component2": {
            "image": "busybox",
            "ports": [ "80/tcp", 81 ]
        },
        "component3": {
            "image": "busybox",
            "ports": "80"
        },
        "component4": {
            "image": "busybox",
            "ports": "80/tcp"
        },
        "component5": {
            "image": "busybox",
            "ports": 8086
        }
    }
}`)

	err := json.Unmarshal(byteSlice, &appDef)
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}
}
