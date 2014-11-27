package userconfig

import (
	"encoding/json"

	"github.com/juju/errgo"
	"labix.org/v2/mgo/bson"
)

func UnmarshalSwarmJson(byteSlice []byte, appConfig *AppConfig) error {
	if err := unmarshal(byteSlice, appConfig); err != nil {
		return Mask(err)
	}

	// If we found arbitrary keys, we need to return an error.
	if len(appConfig.Arbitrary) > 0 {
		// Just return the first invalid field we find.
		for k, _ := range appConfig.Arbitrary {
			// Better reset app-config to its zero value.
			*appConfig = AppConfig{}

			return errgo.WithCausef(nil, ErrUnknownJSONField, "Cannot parse app config. Unknown field '%s' detected.", k)
		}
	}

	return nil
}

func unmarshal(byteSlice []byte, v interface{}) error {
	var j map[string]interface{}

	if err := json.Unmarshal(byteSlice, &j); err != nil {
		return Mask(err)
	}

	byteSlice, err := bson.Marshal(&j)
	if err != nil {
		return Mask(err)
	}

	if err := bson.Unmarshal(byteSlice, v); err != nil {
		return Mask(err)
	}

	return nil
}
