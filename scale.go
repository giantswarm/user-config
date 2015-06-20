package userconfig

import (
	"github.com/juju/errgo"
)

type ScaleDefinition struct {
	// Minimum instances to launch.
	Min int `json:"min,omitempty" description:"Minimum number of instances to launch"`

	// Maximum instances to launch.
	Max int `json:"max,omitempty" description:"Maximum number of instances to launch"`
}

func (sd *ScaleDefinition) validate(valCtx *ValidationContext) error {
	if valCtx == nil {
		return nil
	}

	if sd.Min == 0 {
		sd.Min = valCtx.MinScaleSize
	}

	if sd.Max == 0 {
		sd.Max = valCtx.MaxScaleSize
	}

	if sd.Min < valCtx.MinScaleSize {
		return Mask(errgo.WithCausef(nil, InvalidScalingConfigError, "scale min '%d' cannot be less than '%d'", sd.Min, valCtx.MinScaleSize))
	}

	if sd.Max > valCtx.MaxScaleSize {
		return Mask(errgo.WithCausef(nil, InvalidScalingConfigError, "scale max '%d' cannot be greater than '%d'", sd.Max, valCtx.MaxScaleSize))
	}

	if sd.Min > sd.Max {
		return Mask(errgo.WithCausef(nil, InvalidScalingConfigError, "scale min '%d' cannot be greater than scale max '%d'", sd.Min, sd.Max))
	}

	return nil
}