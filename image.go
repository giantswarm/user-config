package userconfig

import (
	"encoding/json"

	"github.com/giantswarm/generic-types-go"
	"github.com/juju/errgo"
)

type ImageDefinition string

func MustParseImageDefinition(id string) ImageDefinition {
	return ImageDefinition(generictypes.MustParseDockerImage(id).String())
}

func (id ImageDefinition) DockerImage() generictypes.DockerImage {
	return generictypes.MustParseDockerImage(string(id))
}

func (id *ImageDefinition) UnmarshalJSON(data []byte) error {
	var dockerImage generictypes.DockerImage
	if err := json.Unmarshal(data, &dockerImage); err != nil {
		return err
	}

	*id = ImageDefinition(dockerImage.String())

	return nil
}

func (id ImageDefinition) Validate(valCtx *ValidationContext) error {
	if valCtx == nil {
		return nil
	}

	image := id.DockerImage()
	if isGSRegistry(image, valCtx) && image.Namespace != valCtx.Org {
		return Mask(errgo.WithCausef(nil, InvalidImageDefinitionError, "image namespace '%s' must match organization '%s'", image.Namespace, valCtx.Org))
	}

	return nil
}

func isGSRegistry(image generictypes.DockerImage, valCtx *ValidationContext) bool {
	return image.Registry == valCtx.PublicDockerRegistry || image.Registry == valCtx.PrivateDockerRegistry
}
