package userconfig

import (
	"github.com/giantswarm/generic-types-go"
)

type ImageDefinition struct {
	generictypes.DockerImage
}

func MustParseImageDefinition(id string) *ImageDefinition {
	return &ImageDefinition{
		generictypes.MustParseDockerImage(id),
	}
}

func (id ImageDefinition) GenericDockerImage() generictypes.DockerImage {
	return generictypes.MustParseDockerImage(id.String())
}

func (id ImageDefinition) Validate(valCtx *ValidationContext) error {
	if valCtx == nil {
		return nil
	}

	if id.isGSRegistry(valCtx) && id.Namespace != valCtx.Org {
		return maskf(InvalidImageDefinitionError, "image namespace '%s' must match organization '%s'", id.Namespace, valCtx.Org)
	}

	return nil
}

func (id ImageDefinition) isGSRegistry(valCtx *ValidationContext) bool {
	for _, registry := range valCtx.RestrictedRegistries {
		if id.Registry == registry {
			return true
		}
	}
	return false
}
