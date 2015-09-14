package userconfig

import (
	"encoding/json"
	"fmt"
)

type Placement string

const (
	DefaultPlacement       Placement = "simple"
	OnePerMachinePlacement Placement = "one-per-machine"
)

// UnmarshalJSON performs a validation during unmarshaling.
func (pl *Placement) UnmarshalJSON(data []byte) error {
	var s string
	err := json.Unmarshal(data, &s)
	if err != nil {
		return mask(err)
	}

	pv := Placement(s)
	if err := pv.Validate(); err != nil {
		return mask(err)
	}

	*pl = pv
	return nil
}

// Validate checks that the given enum is a valid value.
func (pl Placement) Validate() error {
	switch pl {
	case DefaultPlacement, OnePerMachinePlacement:
	default:
		return maskf(InvalidScalingConfigError, "unknown value for scale placement: '%s'", pl)
	}
	return nil
}

type ScaleDefinition struct {
	// Minimum instances to launch.
	Min int `json:"min,omitempty" description:"Minimum number of instances to launch"`

	// Maximum instances to launch.
	Max int `json:"max,omitempty" description:"Maximum number of instances to launch"`

	Placement Placement `json:"placement,omitempty" description:"Placement strategy when scaling a component. Can be empty or one-per-machine"`
}

func (sd *ScaleDefinition) String() string {
	raw, err := json.Marshal(sd)
	if err != nil {
		panic(fmt.Sprintf("%#v\n", mask(err)))
	}

	return string(raw)
}

func (sd *ScaleDefinition) validate(valCtx *ValidationContext) error {
	if valCtx == nil {
		return nil
	}

	if sd.Min < valCtx.MinScaleSize {
		return maskf(InvalidScalingConfigError, "scale min '%d' cannot be less than '%d'", sd.Min, valCtx.MinScaleSize)
	}

	if sd.Max > valCtx.MaxScaleSize {
		return maskf(InvalidScalingConfigError, "scale max '%d' cannot be greater than '%d'", sd.Max, valCtx.MaxScaleSize)
	}

	if sd.Min > sd.Max {
		return maskf(InvalidScalingConfigError, "scale min '%d' cannot be greater than scale max '%d'", sd.Min, sd.Max)
	}

	if err := sd.Placement.Validate(); err != nil {
		return mask(err)
	}

	return nil
}

func (sd *ScaleDefinition) setDefaults(valCtx *ValidationContext) {
	if sd.Min == 0 {
		sd.Min = valCtx.MinScaleSize
	}

	if sd.Max == 0 {
		sd.Max = valCtx.MaxScaleSize
	}

	if sd.Placement == "" {
		// Allow for valCtx.Placement not be set.
		if valCtx.Placement != "" {
			sd.Placement = valCtx.Placement
		} else {
			sd.Placement = DefaultPlacement
		}
	}
}

func (sd *ScaleDefinition) hideDefaults(valCtx *ValidationContext) *ScaleDefinition {
	if sd.Min == valCtx.MinScaleSize && sd.Max == valCtx.MaxScaleSize {
		return nil
	}

	if sd.Min == valCtx.MinScaleSize {
		sd.Min = 0
	}

	if sd.Max == valCtx.MaxScaleSize {
		sd.Max = 0
	}

	if sd.Placement == DefaultPlacement {
		sd.Placement = ""
	}

	return sd
}

// validateScalingPolicyInPods checks that there all scaling policies within a pod are either not set or the same
func (nds *ComponentDefinitions) validateScalingPolicyInPods() error {
	for componentName, componentDef := range *nds {
		if !componentDef.IsPodRoot() {
			continue
		}

		// Collect all scaling policies
		podComponents, err := nds.PodComponents(componentName)
		if err != nil {
			return mask(err)
		}
		list := []ScaleDefinition{}
		for _, c := range podComponents {
			if c.Scale == nil {
				// No scaling policy set
				continue
			}
			list = append(list, *c.Scale)
		}

		// Check each list for errors
		for i, p1 := range list {
			for j := i + 1; j < len(list); j++ {
				p2 := list[j]
				if p1.Min != 0 && p2.Min != 0 {
					// Both minimums specified, must be the same
					if p1.Min != p2.Min {
						return maskf(InvalidScalingConfigError, "different minimum scaling policies in pod under '%s'", componentName.String())
					}
				}
				if p1.Max != 0 && p2.Max != 0 {
					// Both maximums specified, must be the same
					if p1.Max != p2.Max {
						return maskf(InvalidScalingConfigError, "different maximum scaling policies in pod under '%s'", componentName.String())
					}
				}

				if p1.Placement != "" && p2.Placement != "" {
					if p1.Placement != p2.Placement {
						return maskf(InvalidScalingConfigError, "different scaling placement policies in pod under '%s'", componentName.String())
					}
				}
			}
		}
	}

	// No errors detected
	return nil
}
