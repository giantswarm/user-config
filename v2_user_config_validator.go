package userconfig

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/juju/errgo"
	"github.com/kr/pretty"
)

type v2AppDefCopy V2AppDefinition

func V2CheckForUnknownFields(b []byte, ac *V2AppDefinition) error {
	var clean v2AppDefCopy
	if err := json.Unmarshal(b, &clean); err != nil {
		return Mask(err)
	}

	cleanBytes, err := json.Marshal(clean)
	if err != nil {
		return Mask(err)
	}

	var dirtyMap map[string]interface{}
	if err := json.Unmarshal(b, &dirtyMap); err != nil {
		return Mask(err)
	}
	// Normalize fields to common format
	v2NormalizeEnv(dirtyMap)
	v2NormalizeVolumeSizes(dirtyMap)

	var cleanMap map[string]interface{}
	if err := json.Unmarshal(cleanBytes, &cleanMap); err != nil {
		return Mask(err)
	}

	diffs := pretty.Diff(dirtyMap, cleanMap)
	for _, diff := range diffs {
		*ac = V2AppDefinition{}
		return prettyJSONFieldError(diff)
	}

	return nil
}

func prettyJSONFieldError(diff string) error {
	parts := strings.Split(diff, ":")
	if len(parts) != 2 {
		return errgo.WithCausef(nil, InternalError, "invalid diff format")
	}
	path := parts[0]

	reason := strings.Split(parts[1], "!=")
	if len(parts) != 2 {
		return errgo.WithCausef(nil, InternalError, "invalid diff format")
	}
	missing := strings.Contains(reason[0], "missing")
	unknown := strings.Contains(reason[1], "missing")

	if missing {
		return errgo.WithCausef(nil, MissingJSONFieldError, "missing JSON field: %s", path)
	}

	if unknown {
		return errgo.WithCausef(nil, UnknownJSONFieldError, "unknown JSON field: %s", path)
	}

	return errgo.WithCausef(nil, InternalError, "invalid diff format")
}

// getMapEntry tries to get an entry in the given map that is a string map of
// objects.
func getMapEntry(def map[string]interface{}, key string) map[string]interface{} {
	entry, ok := def[key]
	if !ok {
		// No key element found
		return nil
	}

	entryMap, ok := entry.(map[string]interface{})
	if !ok {
		// entry not right type
		return nil
	}

	return entryMap
}

// v2NormalizeEnv normalizes all struct "env" elements under service/component
// to its natural array format.  This normalization function is expected to
// normalize "valid" data and passthrough everything else.
func v2NormalizeEnv(def map[string]interface{}) {
	nodes := getMapEntry(def, "nodes")
	if nodes == nil {
		// No services element
		return
	}

	for _, node := range nodes {
		nodeMap, ok := node.(map[string]interface{})
		if !ok {
			continue
		}

		envMap := getMapEntry(nodeMap, "env")
		if envMap == nil {
			// No services element
			return
		}

		// 'env' is of map type, normalize it to an array
		// Sort the keys first so the outcome it always the same
		keys := []string{}
		for k, _ := range envMap {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		list := []interface{}{}
		for _, k := range keys {
			v := envMap[k]
			list = append(list, fmt.Sprintf("%s=%s", k, v))
		}
		nodeMap["env"] = list
	}
}

// v2NormalizeVolumeSizes normalizes all volume sizes to it's normalized format
// of "number GB" This normalization function is expected to normalize "valid"
// data and passthrough everything else.
func v2NormalizeVolumeSizes(def map[string]interface{}) {
	nodes := getMapEntry(def, "nodes")
	if nodes == nil {
		// No services element
		return
	}

	for _, node := range nodes {
		nodeMap, ok := node.(map[string]interface{})
		if !ok {
			continue
		}

		volumes := getArrayEntry(nodeMap, "volumes")
		if volumes == nil {
			// No volumes element
			continue
		}

		for _, volume := range volumes {
			volumeMap, ok := volume.(map[string]interface{})
			if !ok {
				continue
			}

			sizeRaw, ok := volumeMap["size"]
			if !ok {
				// Size not found
				continue
			}

			size, ok := sizeRaw.(string)
			if !ok {
				// size is not a string
				continue
			}

			// Parse volume size
			var volumeSize VolumeSize
			// Marshal size string to json
			data, err := json.Marshal(size)
			if err != nil {
				continue
			}

			// Try to unmarshal volume size
			if err := volumeSize.UnmarshalJSON(data); err != nil {
				// Not valid format
				continue
			}
			volumeMap["size"] = string(volumeSize)
		}
	}
}

// validatePods checks that all pods are well formed.
func (nds NodeDefinitions) validatePods() error {
	for name, nodeDef := range nds {
		if nodeDef.Pod == PodChildren || nodeDef.Pod == PodInherit {
			// Check that there are least 2 pod nodes
			children, err := nds.PodNodes(name.String())
			if err != nil {
				return Mask(err)
			}
			if len(children) < 2 {
				return Mask(errgo.WithCausef(nil, InvalidPodConfigError, "Node '%s' must have at least 2 child nodes because if defines 'pod' as '%s'", name, nodeDef.Pod))
			}
			// Children may not have pod set to anything other than empty
			for childName, childDef := range children {
				if childDef.Pod != "" {
					return Mask(errgo.WithCausef(nil, InvalidPodConfigError, "Node '%s' must cannot set 'pod' to '%s' because it is already part of another pod", childName.String(), childDef.Pod))
				}
			}
		}
	}
	return nil
}
