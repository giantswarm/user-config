package userconfig

import (
	"encoding/json"

	"github.com/juju/errgo"
)

// Type of the "pod" field in a V2 node definition.
type PodEnum string

const (
	PodNone     = PodEnum("none")     // No pod is created and no resources are shared.
	PodChildren = PodEnum("children") // A node defining this only configures its direct children to be placed into a pod.
	PodInherit  = PodEnum("inherit")  // A node defining this configures all its children and grand-children to be placed into a pod.
)

// UnmarshalJSON performs a validation during unmarshaling.
func (this *PodEnum) UnmarshalJSON(data []byte) error {
	var s string
	err := json.Unmarshal(data, &s)
	if err != nil {
		return mask(err)
	}

	if s != string(PodNone) && s != string(PodChildren) && s != string(PodInherit) {
		return mask(errgo.WithCausef(nil, InvalidPodConfigError, "Cannot parse app config. Invalid pod value '%s' detected.", s))
	}

	*this = PodEnum(s)
	return nil
}
