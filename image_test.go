package userconfig_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/giantswarm/user-config"
)

func TestUnmarshalInvalidImage(t *testing.T) {
	var newId userconfig.ImageDefinition
	err := json.Unmarshal([]byte(`"foo-bar/image"`), &newId)
	if err == nil {
		t.Fatalf("invalid image not detected")
	}
	if !strings.Contains(err.Error(), "foo-bar") {
		t.Fatalf("expected error to contain invalid image part")
	}
}

func TestImageValidOrgWithPublicRegistry(t *testing.T) {
	valCtx := &userconfig.ValidationContext{
		Org:                   "myorg",
		PublicDockerRegistry:  "registry.giantswarm.io",
		PrivateDockerRegistry: "registry.private.giantswarm.io",
	}

	vd := userconfig.MustParseImageDefinition("registry.giantswarm.io/myorg/foo")
	err := vd.Validate(valCtx)

	if err != nil {
		t.Fatalf("validating image failed: %s", err.Error())
	}
}

func TestImageValidOrgWithPrivateRegistry(t *testing.T) {
	valCtx := &userconfig.ValidationContext{
		Org:                   "myorg",
		PublicDockerRegistry:  "registry.giantswarm.io",
		PrivateDockerRegistry: "registry.private.giantswarm.io",
	}

	vd := userconfig.MustParseImageDefinition("registry.private.giantswarm.io/myorg/foo")
	err := vd.Validate(valCtx)

	if err != nil {
		t.Fatalf("validating image failed: %s", err.Error())
	}
}

func TestImageInvalidOrgWithPublicRegistry(t *testing.T) {
	valCtx := &userconfig.ValidationContext{
		Org:                   "myorg",
		PublicDockerRegistry:  "registry.giantswarm.io",
		PrivateDockerRegistry: "registry.private.giantswarm.io",
	}

	vd := userconfig.MustParseImageDefinition("registry.giantswarm.io/otherorg/foo")
	err := vd.Validate(valCtx)

	if err == nil {
		t.Fatalf("invalid image not detected")
	}
	if !userconfig.IsInvalidImageDefinition(err) {
		t.Fatalf("expected error to be InvalidImageDefinitionError")
	}
}

func TestImageInvalidOrgWithPrivateRegistry(t *testing.T) {
	valCtx := &userconfig.ValidationContext{
		Org:                   "myorg",
		PublicDockerRegistry:  "registry.giantswarm.io",
		PrivateDockerRegistry: "registry.private.giantswarm.io",
	}

	vd := userconfig.MustParseImageDefinition("registry.private.giantswarm.io/otherorg/foo")
	err := vd.Validate(valCtx)

	if err == nil {
		t.Fatalf("invalid image not detected")
	}
	if !userconfig.IsInvalidImageDefinition(err) {
		t.Fatalf("expected error to be InvalidImageDefinitionError")
	}
}
