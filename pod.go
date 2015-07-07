package userconfig

import (
	"encoding/json"
)

// Type of the "pod" field in a node definition.
type PodEnum string

const (
	PodNone     PodEnum = "none"     // No pod is created and no resources are shared.
	PodChildren PodEnum = "children" // A node defining this only configures its direct children to be placed into a pod.
	PodInherit  PodEnum = "inherit"  // A node defining this configures all its children and grand-children to be placed into a pod.
)

// UnmarshalJSON performs a validation during unmarshaling.
func (pe *PodEnum) UnmarshalJSON(data []byte) error {
	var s string
	err := json.Unmarshal(data, &s)
	if err != nil {
		return mask(err)
	}

	pv := PodEnum(s)
	if err := pv.Validate(); err != nil {
		return mask(err)
	}

	*pe = pv
	return nil
}

// Validate checks that the given enum is a valid value.
func (pe PodEnum) Validate() error {
	if pe != PodNone && pe != PodChildren && pe != PodInherit {
		return maskf(InvalidPodConfigError, "invalid pod value '%s'", string(pe))
	}
	return nil
}
