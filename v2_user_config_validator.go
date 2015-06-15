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

	diff := pretty.Diff(dirtyMap, cleanMap)
	for _, v := range diff {
		*ac = V2AppDefinition{}

		field := strings.Split(v, ":")
		return errgo.WithCausef(nil, ErrUnknownJSONField, "Cannot parse app definition. Unknown field '%s' detected.", field[0])
	}

	return nil
}

// v2NormalizeEnv normalizes all struct "env" elements under service/component
// to its natural array format.  This normalization function is expected to
// normalize "valid" data and passthrough everything else.
func v2NormalizeEnv(def map[string]interface{}) {
	nodes := getArrayEntry(def, "nodes")
	if nodes == nil {
		// No services element
		return
	}

	for _, node := range nodes {
		nodeMap, ok := node.(map[string]interface{})
		if !ok {
			continue
		}

		env, ok := nodeMap["env"]
		if !ok {
			continue
		}

		envMap, ok := env.(map[string]interface{})
		if !ok {
			// Not of the map type
			continue
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
	nodes := getArrayEntry(def, "nodes")
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

// validate performs semantic validations of this V2AppDefinition.
// Return the first possible error.
func (ad *V2AppDefinition) validate() error {
	for _, n := range ad.Nodes {
		if err := n.validate(); err != nil {
			return err
		}
	}

	return nil
}

// validate performs semantic validations of this NodeDefinition.
// Return the first possible error.
func (nd *NodeDefinition) validate() error {
	// Detect duplicate volume "path"
	paths := make(map[string]string)
	for _, v := range nd.Volumes {
		path := v.Path
		if _, found := paths[path]; found {
			return errgo.WithCausef(nil, ErrDuplicateVolumePath, "Cannot parse app definition. Duplicate volume '%s' detected.", path)
		}
		paths[path] = path
	}

	for d, _ := range nd.Domains {
		if err := d.Validate(); err != nil {
			return Mask(err)
		}
	}

	// No errors found
	return nil
}
