package userconfig_test

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/giantswarm/user-config"
)

func TestV2AppLinksScaleDefaults(t *testing.T) {
	a := ExampleDefinition()
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
	a := ExampleDefinition()
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

	a := ExampleDefinition()
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

	a := ExampleDefinition()
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

func TestV2ScaleAppHidePlacement(t *testing.T) {
	a := ExampleDefinition()
	a.Components["component/a"].Scale = &userconfig.ScaleDefinition{
		Placement: "simple",
	}

	valCtx := NewValidationContext()
	valCtx.Placement = "simple"

	err := a.SetDefaults(valCtx)
	if err != nil {
		t.Fatalf("setting defaults failed: %s", err.Error())
	}
	if a.Components["component/a"].Scale.Placement != valCtx.Placement {
		t.Fatalf("expected placement scale to be '%s'", valCtx.Placement)
	}

	b, err := a.HideDefaults(valCtx)
	if err != nil {
		t.Fatalf("hiding defaults failed: %s", err.Error())
	}

	if b.Components["component/a"].Scale != nil {
		t.Fatalf("scale should be hidden")
	}
}

func TestV2ScaleAppDontHideCustomPlacement(t *testing.T) {
	customPlacement := userconfig.Placement("one-per-machine")
	a := ExampleDefinition()
	a.Components["component/a"].Scale = &userconfig.ScaleDefinition{
		Placement: customPlacement,
	}

	valCtx := NewValidationContext()
	valCtx.Placement = "simple"

	err := a.SetDefaults(valCtx)
	if err != nil {
		t.Fatalf("setting defaults failed: %s", err.Error())
	}
	if a.Components["component/a"].Scale.Placement != customPlacement {
		t.Fatalf("expected placement scale to be '%s'", customPlacement)
	}

	b, err := a.HideDefaults(valCtx)
	if err != nil {
		t.Fatalf("hiding defaults failed: %s", err.Error())
	}

	if b.Components["component/a"].Scale == nil {
		t.Fatalf("scale should not be hidden")
	}

	if b.Components["component/a"].Scale.Placement != customPlacement {
		t.Fatalf("expected placement scale to be '%s'", customPlacement)
	}
}

func TestValidPlacementValues(t *testing.T) {
	list := []string{
		"simple",
		"one-per-machine",
	}

	for _, s := range list {
		var pe struct {
			Placement userconfig.Placement
		}
		data := fmt.Sprintf(`{"placement": "%s"}`, s)
		if err := json.Unmarshal([]byte(data), &pe); err != nil {
			t.Fatalf("Valid placement value '%s' considered invalid because %v", s, err)
		}
	}
}

func TestInvalidPlacementValues(t *testing.T) {
	list := []string{
		"foo",
		"none",
		"two-per-machine",
		"",
	}

	for _, s := range list {
		var pe struct {
			Placement userconfig.Placement
		}
		data := fmt.Sprintf(`{"placement": "%s"}`, s)
		if err := json.Unmarshal([]byte(data), &pe); err == nil {
			t.Fatalf("Invalid placement value '%s' considered valid", s)
		}
	}
}
