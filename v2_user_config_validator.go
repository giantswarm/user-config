package userconfig

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/giantswarm/generic-types-go"
	"github.com/kr/pretty"
)

type v2AppDefCopy V2AppDefinition

func V2CheckForUnknownFields(b []byte, ac *V2AppDefinition) error {
	var clean v2AppDefCopy
	if err := json.Unmarshal(b, &clean); err != nil {
		return mask(err)
	}

	cleanBytes, err := json.Marshal(clean)
	if err != nil {
		return mask(err)
	}

	var dirtyMap map[string]interface{}
	if err := json.Unmarshal(b, &dirtyMap); err != nil {
		return mask(err)
	}

	// Normalize fields to common format
	v2NormalizeEnv(dirtyMap)
	v2NormalizeDomains(dirtyMap)
	v2NormalizeVolumeSizes(dirtyMap)
	v2NormalizePorts(dirtyMap)

	var cleanMap map[string]interface{}
	if err := json.Unmarshal(cleanBytes, &cleanMap); err != nil {
		return mask(err)
	}

	// Also normalize the clean map for ports because marshalling
	// ports preserves the input format
	v2NormalizeDomains(cleanMap)
	v2NormalizePorts(cleanMap)

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
		return maskf(InternalError, "invalid diff format '%s'", diff)
	}
	path := parts[0]

	reason := strings.Split(parts[1], "!=")
	if len(parts) != 2 {
		return maskf(InternalError, "invalid diff format '%s'", diff)
	}
	missing := strings.Contains(reason[0], "missing")
	unknown := strings.Contains(reason[1], "missing")

	if missing {
		return maskf(MissingJSONFieldError, "missing JSON field: %s", path)
	}

	if unknown {
		return maskf(UnknownJSONFieldError, "unknown JSON field: %s", path)
	}

	return maskf(WrongDiffOrderError, "wrong diff order: %s", strings.Trim(parts[1], " "))
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
// to its natural array format. This normalization function is expected to
// normalize "valid" data and passthrough everything else.
func v2NormalizeEnv(def map[string]interface{}) {
	components := getMapEntry(def, "components")
	if components == nil {
		// No components element
		return
	}

	for _, component := range components {
		componentMap, ok := component.(map[string]interface{})
		if !ok {
			continue
		}

		envMap := getMapEntry(componentMap, "env")
		if envMap == nil {
			// No env field formatted as map
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
		componentMap["env"] = list
	}
}

// v2NormalizeDomains normalizes all domain objects to adhere to the
// `port: domainList` format
func v2NormalizeDomains(def map[string]interface{}) {
	components := getMapEntry(def, "components")
	if components == nil {
		// No services element
		return
	}

	for _, component := range components {
		componentMap, ok := component.(map[string]interface{})
		if !ok {
			continue
		}

		domainMap := getMapEntry(componentMap, "domains")
		if domainMap == nil {
			// No domains element
			continue
		}

		// If the keys are not ports, reverse the map
		newMap := make(map[string]interface{})
		for k, v := range domainMap {
			var newKey string
			var newValue interface{}
			// Try to unmarshal the key as a port
			if port, err := generictypes.ParseDockerPort(k); err != nil {
				// Key is not a port, assume it is a domain
				// value should be a port then
				if portStr, ok := v.(string); ok {
					// Parse the port
					if port, err := generictypes.ParseDockerPort(portStr); err != nil {
						// It is not a valid port, give up
						continue
					} else {
						newKey = port.String()
						newValue = k
					}
				} else if portFloat, ok := v.(float64); ok {
					// key: domain, value: port (int)
					portInt := strconv.FormatFloat(portFloat, 'f', 0, 64)
					if port, err := generictypes.ParseDockerPort(portInt); err != nil {
						// unknown format, this should not happen since validation should prevent that
						continue
					} else {
						newKey = port.String()
						newValue = k
					}
				} else {
					// unknown format, this should not happen since validation should prevent that
					continue
				}
			} else {
				// Key is a port, keep the value
				newKey = port.String()
				newValue = sortStringSlice(v)
			}

			if existingValue, ok := newMap[newKey]; ok {
				// Key already has a value, append to it
				newMap[newKey] = sortStringSlice(appendInterfaceList(existingValue, newValue))
			} else {
				// Key has no value yet, create it
				newMap[newKey] = sortStringSlice(appendInterfaceList(nil, newValue))
			}
		}

		// 'domains' is of map type, normalize it to an array
		// Sort the keys first so the outcome it always the same
		keys := []string{}
		for k, _ := range newMap {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		sortedMap := make(map[string]interface{})
		for _, k := range keys {
			sortedMap[k] = newMap[k]
		}
		componentMap["domains"] = sortedMap
	}
}

// appendInterfaceList appends given value to given list.
// list is assume to be of type []interface{}, value can be an array
// or a single value
func appendInterfaceList(list interface{}, value interface{}) interface{} {
	valueAsList, ok := value.([]interface{})
	if !ok {
		valueAsList = []interface{}{value}
	}

	if list == nil {
		return valueAsList
	}
	if listAsList, ok := list.([]interface{}); !ok {
		// This should not happen.
		// Return list as this will trigger an error in the comparison phase
		return list
	} else {
		return append(listAsList, valueAsList...)
	}
}

// sortStringSlice sorts the given value if it is a string slice.
// if not the value is returned unmodified
func sortStringSlice(value interface{}) interface{} {
	if interfaceSlice, ok := value.([]interface{}); ok {
		// sort the string slice
		stringSlice := []string{}
		for _, v := range interfaceSlice {
			if s, ok := v.(string); ok {
				stringSlice = append(stringSlice, s)
			}
		}
		sort.Strings(stringSlice)

		// move sorted values back to an interface slice
		newInterfaceSlice := []interface{}{}
		for _, s := range stringSlice {
			newInterfaceSlice = append(newInterfaceSlice, s)
		}

		return newInterfaceSlice
	} else {
		return value
	}
}

// v2NormalizePorts normalizes all values of "ports" elements
func v2NormalizePorts(def map[string]interface{}) {
	components := getMapEntry(def, "components")
	if components == nil {
		// No components element
		return
	}

	for _, component := range components {
		componentMap, ok := component.(map[string]interface{})
		if !ok {
			continue
		}

		portsRaw, ok := componentMap["ports"]
		if !ok {
			// No ports field
			continue
		}

		// Check for string value
		if portsStr, ok := portsRaw.(string); ok {
			// Port is a string value, wrap it in an array
			componentMap["ports"] = []interface{}{normalizeSinglePort(portsStr)}
			continue
		}

		// Check for number values
		if portsNum, ok := portsRaw.(float64); ok {
			// Port is a number
			componentMap["ports"] = []interface{}{normalizeSinglePort(portsNum)}
			continue
		}

		// Check for array values
		if portsArr, ok := portsRaw.([]interface{}); ok {
			// Ports is an array, normalize all elements
			for i, v := range portsArr {
				portsArr[i] = normalizeSinglePort(v)
			}
			componentMap["ports"] = portsArr
			continue
		}
	}
}

func normalizeSinglePort(input interface{}) interface{} {
	// Check for string value
	if str, ok := input.(string); ok {
		// Port is a string value, parse it
		if port, err := generictypes.ParseDockerPort(str); err == nil {
			return port.String()
		}
		// Error, just return input
		return input
	}

	// Check for number values
	if num, ok := input.(float64); ok {
		// Port is a number, parse it
		if port, err := generictypes.ParseDockerPort(strconv.Itoa(int(num))); err == nil {
			return port.String()
		}
		// Error, just return input
		return input
	}

	// Unknown type, just return input
	return input
}

// v2NormalizeVolumeSizes normalizes all volume sizes to it's normalized format
// of "number GB" This normalization function is expected to normalize "valid"
// data and passthrough everything else.
func v2NormalizeVolumeSizes(def map[string]interface{}) {
	components := getMapEntry(def, "components")
	if components == nil {
		// No services element
		return
	}

	for _, component := range components {
		componentMap, ok := component.(map[string]interface{})
		if !ok {
			continue
		}

		volumes := getArrayEntry(componentMap, "volumes")
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
func (nds ComponentDefinitions) validatePods() error {
	for name, componentDef := range nds {
		if componentDef.Pod == PodChildren || componentDef.Pod == PodInherit {
			// Check that there are least 2 pod components
			children, err := nds.PodComponents(name)
			if err != nil {
				return mask(err)
			}
			if len(children) < 2 {
				return maskf(InvalidPodConfigError, "component '%s' must have at least 2 child components because if defines 'pod' as '%s'", name, componentDef.Pod)
			}
			// Children may not have pod set to anything other than empty
			for childName, childDef := range children {
				if childDef.Pod != "" {
					return maskf(InvalidPodConfigError, "component '%s' must cannot set 'pod' to '%s' because it is already part of another pod", childName.String(), childDef.Pod)
				}
			}
		}
	}
	return nil
}

// validateLeafs checks that all leaf components are a component.
func (nds ComponentDefinitions) validateLeafs() error {
	for componentName, componentDef := range nds {
		if nds.IsLeaf(componentName) {
			// It has to be a component
			if !componentDef.IsComponent() {
				return maskf(InvalidComponentDefinitionError, "component '%s' must have an 'image'", componentName.String())
			}
		}
	}
	return nil
}

// getArrayEntry tries to get an entry in the given map that is an array of
// objects.
func getArrayEntry(def map[string]interface{}, key string) []interface{} {
	entry, ok := def[key]
	if !ok {
		// No key element found
		return nil
	}

	entryArr, ok := entry.([]interface{})
	if !ok {
		// entry not right type
		return nil
	}

	return entryArr
}

// normalizeFolder removes any trailing path separator from the given path.
func normalizeFolder(path string) string {
	if path == "" {
		return ""
	}
	l := len(path)
	if os.IsPathSeparator(path[l-1]) {
		path = path[:l-1]
	}
	return path
}
