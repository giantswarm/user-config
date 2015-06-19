package userconfig_test

import (
	"encoding/json"
	"testing"

	"github.com/giantswarm/user-config"
)

func TestV2AppLinksScaleDefaults(t *testing.T) {
	a := V2ExampleDefinition()
	raw, err := json.Marshal(a)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}

	var b userconfig.V2AppDefinition
	err = json.Unmarshal(raw, &b)
	if err != nil {
		t.Fatalf("json.Unmarshal failed: %s", err.Error())
	}
	if b.Nodes["node/a"].Scale.Min != userconfig.MinScaleSize {
		t.Fatalf("expetced default min scale to be '%d'", userconfig.MinScaleSize)
	}
	if b.Nodes["node/a"].Scale.Max != userconfig.MaxScaleSize {
		t.Fatalf("expetced default max scale to be '%d'", userconfig.MaxScaleSize)
	}
}
