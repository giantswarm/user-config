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

func TestTypeAssertDockerImage(t *testing.T) {
	id := userconfig.MustParseImageDefinition("registry.giantswarm.io/myorg/foo")
	dockerImage := id.GenericDockerImage()
	if dockerImage.Registry != "registry.giantswarm.io" {
		t.Fatalf("failed to type assert generictypes.DockerImage")
	}
}

func TestImageValidOrgWithPublicRegistry(t *testing.T) {
	valCtx := &userconfig.ValidationContext{
		Org:                   "myorg",
		PublicDockerRegistry:  "registry.giantswarm.io",
		PrivateDockerRegistry: "registry.private.giantswarm.io",
	}

	id := userconfig.MustParseImageDefinition("registry.giantswarm.io/myorg/foo")
	err := id.Validate(valCtx)

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

	id := userconfig.MustParseImageDefinition("registry.private.giantswarm.io/myorg/foo")
	err := id.Validate(valCtx)

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

	id := userconfig.MustParseImageDefinition("registry.giantswarm.io/otherorg/foo")
	err := id.Validate(valCtx)

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

	id := userconfig.MustParseImageDefinition("registry.private.giantswarm.io/otherorg/foo")
	err := id.Validate(valCtx)

	if err == nil {
		t.Fatalf("invalid image not detected")
	}
	if !userconfig.IsInvalidImageDefinition(err) {
		t.Fatalf("expected error to be InvalidImageDefinitionError")
	}
}
