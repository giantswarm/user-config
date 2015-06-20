package userconfig_test

import (
	"testing"

	"github.com/giantswarm/user-config"
)

func TestImageValidOrgWithPublicRegistry(t *testing.T) {
	valCtx := &userconfig.ValidationContext{
		Org:                   "myorg",
		PublicDockerRegistry:  "registry.giantswarm.io",
		PrivateDockerRegistry: "registry.private.giantswarm.io",
	}

	vd := userconfig.NewImageDefinition("registry.giantswarm.io/myorg/foo")
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

	vd := userconfig.NewImageDefinition("registry.private.giantswarm.io/myorg/foo")
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

	vd := userconfig.NewImageDefinition("registry.giantswarm.io/otherorg/foo")
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

	vd := userconfig.NewImageDefinition("registry.private.giantswarm.io/otherorg/foo")
	err := vd.Validate(valCtx)

	if err == nil {
		t.Fatalf("invalid image not detected")
	}
	if !userconfig.IsInvalidImageDefinition(err) {
		t.Fatalf("expected error to be InvalidImageDefinitionError")
	}
}
