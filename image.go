package userconfig

import (
	"github.com/giantswarm/generic-types-go"
	"github.com/juju/errgo"
)

type ImageDefinition struct {
	generictypes.DockerImage
}

func MustParseImageDefinition(id string) ImageDefinition {
	return ImageDefinition{
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
		return Mask(errgo.WithCausef(nil, InvalidImageDefinitionError, "image namespace '%s' must match organization '%s'", id.Namespace, valCtx.Org))
	}

	return nil
}

func (id ImageDefinition) isGSRegistry(valCtx *ValidationContext) bool {
	return id.Registry == valCtx.PublicDockerRegistry || id.Registry == valCtx.PrivateDockerRegistry
}
