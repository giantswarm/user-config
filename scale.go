package userconfig

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

	if sd.Min < valCtx.MinScaleSize {
		return maskf(InvalidScalingConfigError, "scale min '%d' cannot be less than '%d'", sd.Min, valCtx.MinScaleSize)
	}

	if sd.Max > valCtx.MaxScaleSize {
		return maskf(InvalidScalingConfigError, "scale max '%d' cannot be greater than '%d'", sd.Max, valCtx.MaxScaleSize)
	}

	if sd.Min > sd.Max {
		return maskf(InvalidScalingConfigError, "scale min '%d' cannot be greater than scale max '%d'", sd.Min, sd.Max)
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

	return sd
}

// validateScalingPolicyInPods checks that there all scaling policies within a pod are either not set of the same
func (nds *NodeDefinitions) validateScalingPolicyInPods() error {
	for nodeName, nodeDef := range *nds {
		if !nodeDef.IsPodRoot() {
			continue
		}

		// Collect all scaling policies
		podNodes, err := nds.PodNodes(nodeName.String())
		if err != nil {
			return mask(err)
		}
		list := []ScaleDefinition{}
		for _, c := range podNodes {
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
						return maskf(InvalidScalingConfigError, "Cannot parse app config. Different minimum scaling policies in pod under '%s'.", nodeName.String())
					}
				}
				if p1.Max != 0 && p2.Max != 0 {
					// Both maximums specified, must be the same
					if p1.Max != p2.Max {
						return maskf(InvalidScalingConfigError, "Cannot parse app config. Different maximum scaling policies in pod under '%s'.", nodeName.String())
					}
				}
			}
		}
	}

	// No errors detected
	return nil
}
