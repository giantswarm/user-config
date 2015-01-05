package userconfig

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/juju/errgo"
	"github.com/kr/pretty"
)

type appConfigCopy AppConfig

func CheckForUnknownFields(b []byte, ac *AppConfig) error {
	var clean appConfigCopy
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
	normalizeEnv(dirtyMap)

	var cleanMap map[string]interface{}
	if err := json.Unmarshal(cleanBytes, &cleanMap); err != nil {
		return Mask(err)
	}

	diff := pretty.Diff(dirtyMap, cleanMap)
	for _, v := range diff {
		*ac = AppConfig{}

		field := strings.Split(v, ":")
		return errgo.WithCausef(nil, ErrUnknownJSONField, "Cannot parse app config. Unknown field '%s' detected.", field[0])
	}

	return nil
}

// normalizeEnv normalizes all struct "env" elements under service/component to its natural array format.
func normalizeEnv(config map[string]interface{}) {
	services := getArrayEntry(config, "services")
	if services == nil {
		// No services element
		return
	}
	for _, service := range services {
		serviceMap, ok := service.(map[string]interface{})
		if !ok {
			continue
		}
		components := getArrayEntry(serviceMap, "components")
		if components == nil {
			// No components element
			continue
		}
		for _, component := range components {
			componentMap, ok := component.(map[string]interface{})
			if !ok {
				continue
			}
			env, ok := componentMap["env"]
			if !ok {
				continue
			}
			envMap, ok := env.(map[string]interface{})
			if !ok {
				// Not of the map type
				continue
			}
			// 'env' is of map type, normalize it to an array
			list := []interface{}{}
			for k, v := range envMap {
				list = append(list, fmt.Sprintf("%s=%s", k, v))
			}
			componentMap["env"] = list
		}
	}
}

// getArrayEntry tries to get an entry in the given map that is an array of objects.
func getArrayEntry(config map[string]interface{}, key string) []interface{} {
	entry, ok := config[key]
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
