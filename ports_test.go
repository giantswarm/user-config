package userconfig_test

import (
	"encoding/json"
	"testing"

	"github.com/giantswarm/user-config"
)

func TestValidPortsValues(t *testing.T) {
	list := []string{
		`80`,
		`"80/tcp"`,
		`["80/tcp","81/tcp"]`,
		`[80,"81/tcp","82"]`,
	}

	for _, s := range list {
		var pds userconfig.PortDefinitions
		if err := json.Unmarshal([]byte(s), &pds); err != nil {
			t.Fatalf("Valid port definitions value '%s' considered invalid because %v", s, err)
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
