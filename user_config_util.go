package userconfig

import (
	"encoding/json"
	"reflect"
	"regexp"
	"strings"
)

var (
	lowerRegex = regexp.MustCompile("[a-z]")
	upperRegex = regexp.MustCompile("[A-Z]")
)

// input: "serviceName", "session",                 output: "serviceName"     "session"
// input: "serviceName", "session-service/session", output: "session-service" "session"
func ParseDependency(serviceName, dependency string) (string, string) {
	slashSplitted := strings.Split(dependency, "/")

	depServiceName := ""
	depComponentName := ""

	if len(slashSplitted) == 1 {
		depServiceName = serviceName
		depComponentName = dependency
	} else {
		depServiceName = slashSplitted[0]
		depComponentName = slashSplitted[1]
	}

	return depServiceName, depComponentName
}

// FixJSONFieldNames expects an byte array representing a valid JSON string.
// The given JSON field names will be transformed from upper cased to
// underscore.
func FixJSONFieldNames(b []byte) ([]byte, error) {
	var j map[string]interface{}

	if err := json.Unmarshal(b, &j); err != nil {
		return nil, Mask(err)
	}

	j = fixJSONFieldNamesRecursive(j)

	b, err := json.Marshal(j)
	if err != nil {
		return nil, Mask(err)
	}

	return b, nil
}

// fixJSONFieldNamesRecursive transforms the keys of the given map from
// uppercased to underscore.
func fixJSONFieldNamesRecursive(j map[string]interface{}) map[string]interface{} {
	for k, v := range j {
		delete(j, k)
		k = FixFieldName(k)
		j[k] = v

		if reflect.TypeOf(v).Kind() == reflect.Map {
			if m, ok := v.(map[string]interface{}); ok {
				j[k] = fixJSONFieldNamesRecursive(m)
			}

			continue
		}

		if reflect.TypeOf(v).Kind() == reflect.Slice {
			if s, ok := v.([]interface{}); ok {
				for i, item := range s {
					m := map[string]interface{}{
						k: item,
					}

					m = fixJSONFieldNamesRecursive(m)
					s[i] = m[k]
				}

				j[k] = s
			}

			continue
		}
	}

	return j
}

// FixFieldName transforms upper cased strings into underscore ones.
//
//   "appName"       => "app_name"
//   "Services"      => "services"
//   "ComponentName" => "component_name"
//
func FixFieldName(s string) string {
	r := strings.Split(s, "")

	for i, v := range r {
		isUpper := upperRegex.Match([]byte(v))
		if isUpper {
			r[i] = strings.ToLower(v)
		}

		if i == 0 {
			continue
		}

		needFix := lowerRegex.Match([]byte(r[i-1]))
		if needFix && isUpper {
			r[i] = "_" + r[i]
		}
	}

	return strings.Join(r, "")
}
