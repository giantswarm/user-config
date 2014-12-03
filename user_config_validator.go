package userconfig

import (
	"encoding/json"
	"github.com/kr/pretty"
	"strings"

	"github.com/juju/errgo"
)

func CheckForUnknownFields(b []byte, ac *AppConfig) error {
	var clean AppConfigCopy
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
