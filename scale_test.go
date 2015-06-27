package userconfig_test

import (
	"testing"
)

func TestV2AppLinksScaleDefaults(t *testing.T) {
	a := V2ExampleDefinition()
	valCtx := NewValidationContext()

	err := a.SetDefaults(valCtx)
	if err != nil {
		t.Fatalf("validation failed: %s", err.Error())
	}
	if a.Nodes["node/a"].Scale.Min != valCtx.MinScaleSize {
		t.Fatalf("expected default min scale to be '%d'", valCtx.MinScaleSize)
	}
	if a.Nodes["node/a"].Scale.Max != valCtx.MaxScaleSize {
		t.Fatalf("expected default max scale to be '%d'", valCtx.MaxScaleSize)
	}
}
