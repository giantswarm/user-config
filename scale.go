package userconfig

import (
	"github.com/juju/errgo"
)

var (
	defaultMinScaleSize = 1
	MinScaleSize        = 0
	defaultMaxScaleSize = 10
	MaxScaleSize        = 0
)

func init() {
	SetDefaultMinScaleSize()
	SetDefaultMaxScaleSize()
}

func SetDefaultMinScaleSize() {
	MinScaleSize = defaultMinScaleSize
}

func SetMinScaleSize(min int) {
	MinScaleSize = min
}

func SetDefaultMaxScaleSize() {
	MaxScaleSize = defaultMaxScaleSize
}

func SetMaxScaleSize(max int) {
	MaxScaleSize = max
}

type ScaleDefinition struct {
	// Minimum instances to launch.
	Min int `json:"min,omitempty" description:"Minimum number of instances to launch"`

	// Maximum instances to launch.
	Max int `json:"max,omitempty" description:"Maximum number of instances to launch"`
}

func (sd ScaleDefinition) validate() error {
	if sd.Min < MinScaleSize {
		return Mask(errgo.WithCausef(nil, InvalidScalingConfigError, "scale min '%d' cannot be less than '%d'", sd.Min, MinScaleSize))
	}

	if sd.Max > MaxScaleSize {
		return Mask(errgo.WithCausef(nil, InvalidScalingConfigError, "scale max '%d' cannot be greater than '%d'", sd.Max, MaxScaleSize))
	}

	if sd.Min > sd.Max {
		return Mask(errgo.WithCausef(nil, InvalidScalingConfigError, "scale min '%d' cannot be greater than scale max '%d'", sd.Min, sd.Max))
	}

	return nil
}
