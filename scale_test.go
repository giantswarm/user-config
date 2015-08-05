package userconfig_test

import (
	"testing"

	"github.com/giantswarm/user-config"
)

func TestV2AppLinksScaleDefaults(t *testing.T) {
	a := V2ExampleDefinition()
	valCtx := NewValidationContext()

	err := a.SetDefaults(valCtx)
	if err != nil {
		t.Fatalf("validation failed: %s", err.Error())
	}
	if a.Components["component/a"].Scale.Min != valCtx.MinScaleSize {
		t.Fatalf("expected default min scale to be '%d'", valCtx.MinScaleSize)
	}
	if a.Components["component/a"].Scale.Max != valCtx.MaxScaleSize {
		t.Fatalf("expected default max scale to be '%d'", valCtx.MaxScaleSize)
	}
}

func TestV2AppScaleHideDefaults(t *testing.T) {
	a := V2ExampleDefinition()
	valCtx := NewValidationContext()
	valCtx.MinScaleSize = 3
	valCtx.MaxScaleSize = 8

	err := a.SetDefaults(valCtx)
	if err != nil {
		t.Fatalf("setting defaults failed: %s", err.Error())
	}
	if a.Components["component/a"].Scale.Min != valCtx.MinScaleSize {
		t.Fatalf("expected default min scale to be '%d'", valCtx.MinScaleSize)
	}
	if a.Components["component/a"].Scale.Max != valCtx.MaxScaleSize {
		t.Fatalf("expected default max scale to be '%d'", valCtx.MaxScaleSize)
	}

	b, err := a.HideDefaults(valCtx)
	if err != nil {
		t.Fatalf("hiding defaults failed: %s", err.Error())
	}

	if b.Components["component/a"].Scale != nil {
		t.Fatalf("scale not hidden")
	}
}

// TestV2AppScaleHideMinScale ensures that min scale is hidden when marshalled,
// in case the user did not actively change the internal default value.
func TestV2AppScaleHideMinScale(t *testing.T) {
	customScale := 6

	a := V2ExampleDefinition()
	a.Components["component/a"].Scale = &userconfig.ScaleDefinition{
		Max: customScale,
	}

	valCtx := NewValidationContext()
	valCtx.MinScaleSize = 3
	valCtx.MaxScaleSize = 8

	err := a.SetDefaults(valCtx)
	if err != nil {
		t.Fatalf("setting defaults failed: %s", err.Error())
	}
	if a.Components["component/a"].Scale.Min != valCtx.MinScaleSize {
		t.Fatalf("expected default min scale to be '%d'", valCtx.MinScaleSize)
	}
	if a.Components["component/a"].Scale.Max != customScale {
		t.Fatalf("expected max scale to be '%d'", customScale)
	}

	b, err := a.HideDefaults(valCtx)
	if err != nil {
		t.Fatalf("hiding defaults failed: %s", err.Error())
	}

	if b.Components["component/a"].Scale == nil {
		t.Fatalf("scale hidden")
	}

	if b.Components["component/a"].Scale.Min != 0 {
		t.Fatalf("min scale NOT hidden")
	}
}

// TestV2AppScaleHideMaxScale ensures that max scale is hidden when marshalled,
// in case the user did not actively change the internal default value.
func TestV2AppScaleHideMaxScale(t *testing.T) {
	customScale := 6

	a := V2ExampleDefinition()
	a.Components["component/a"].Scale = &userconfig.ScaleDefinition{
		Min: customScale,
	}

	valCtx := NewValidationContext()
	valCtx.MinScaleSize = 3
	valCtx.MaxScaleSize = 8

	err := a.SetDefaults(valCtx)
	if err != nil {
		t.Fatalf("setting defaults failed: %s", err.Error())
	}
	if a.Components["component/a"].Scale.Min != customScale {
		t.Fatalf("expected custom min scale to be '%d'", customScale)
	}
	if a.Components["component/a"].Scale.Max != valCtx.MaxScaleSize {
		t.Fatalf("expected default max scale to be '%d'", valCtx.MaxScaleSize)
	}

	b, err := a.HideDefaults(valCtx)
	if err != nil {
		t.Fatalf("hiding defaults failed: %s", err.Error())
	}

	if b.Components["component/a"].Scale == nil {
		t.Fatalf("scale hidden")
	}

	if b.Components["component/a"].Scale.Max != 0 {
		t.Fatalf("max scale NOT hidden")
	}
}
