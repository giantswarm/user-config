package userconfig_test

import (
	"testing"

	"github.com/giantswarm/user-config"
)

func TestV2AppLinksScaleDefaults(t *testing.T) {
	a := V2ExampleDefinition()

	valCtx := &userconfig.ValidationContext{
		MinScaleSize: 1,
		MaxScaleSize: 10,
	}

	err := a.Validate(valCtx)
	if err != nil {
		t.Fatalf("validation failed: %s", err.Error())
	}
	if a.Nodes["node/a"].Scale.Min != valCtx.MinScaleSize {
		t.Fatalf("expetced default min scale to be '%d'", valCtx.MinScaleSize)
	}
	if a.Nodes["node/a"].Scale.Max != valCtx.MaxScaleSize {
		t.Fatalf("expetced default max scale to be '%d'", valCtx.MaxScaleSize)
	}
}
