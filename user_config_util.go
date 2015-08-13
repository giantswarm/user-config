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

// FixJSONFieldNames expects an byte array representing a valid JSON string.
// The given JSON field names will be transformed from upper cased to
// underscore.
func FixJSONFieldNames(b []byte) ([]byte, error) {
	var j map[string]interface{}

	if err := json.Unmarshal(b, &j); err != nil {
		return nil, mask(err)
	}

	j = fixJSONFieldNamesRecursive(j, "")

	b, err := json.Marshal(j)
	if err != nil {
		return nil, mask(err)
	}

	return b, nil
}

// fixJSONFieldNamesRecursive transforms the keys of the given map from
// uppercased to underscore.
func fixJSONFieldNamesRecursive(j map[string]interface{}, keyPrefix string) map[string]interface{} {
	// Exclude some keys from fixing. This also excludes all keys that are
	// eventually stored under that given key.
	if keyPrefix == "/services/components/env" {
		return j
	}

	// Exclude V2 env fields
	keyPrefixParts := strings.Split(keyPrefix, "/")
	if strings.HasPrefix(keyPrefix, "/components/") && keyPrefixParts[len(keyPrefixParts)-1] == "env" {
		// We found a V2 "env" field
		return j
	}

	// Exclude some keys from fixing. This also excludes all keys that are
	// eventually stored under that given key.
	if keyPrefix == "/services/components/domains" {
		return j
	}

	if strings.HasPrefix(keyPrefix, "/components/") && keyPrefixParts[len(keyPrefixParts)-1] == "domains" {
		// We found a V2 "env" field
		return j
	}

	for k, v := range j {
		// This is really tricky. Component names must be arbitrary strings. We are not
		// allowed to fix them. Further everything inside the components needs to be
		// fixed. So we just ignore the component names itself.
		if keyPrefix != "/components" {
			delete(j, k)
			k = FixFieldName(k)
			j[k] = v
		}

		if v == nil {
			continue
		}

		if reflect.TypeOf(v).Kind() == reflect.Map {
			if m, ok := v.(map[string]interface{}); ok {
				j[k] = fixJSONFieldNamesRecursive(m, keyPrefix+"/"+k)
			}

			continue
		}

		if reflect.TypeOf(v).Kind() == reflect.Slice {
			if s, ok := v.([]interface{}); ok {
				for i, item := range s {
					m := map[string]interface{}{
						k: item,
					}

					m = fixJSONFieldNamesRecursive(m, keyPrefix)
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
