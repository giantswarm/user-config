package userconfig

import (
	"encoding/json"

	"github.com/juju/errgo"
	"labix.org/v2/mgo/bson"
)

// CheckForUnknownFields looks up AppConfig.Arbitrary, that contains arbitrary
// keys, if any. Those arbitrary keys are handled as errors by us, since we
// want to improve user feedback.
func CheckForUnknownFields(appConfig *AppConfig) error {
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

// UnmarshalWithBSONUnmarshaler unmarshals data given by byteSlice into the
// given pointer interface. The BSON unmarshaler is special in terms of
// detecting arbitrary JSON fields, that can be inlined in the given pointer
// interface. For more information see http://godoc.org/labix.org/v2/mgo/bson.
func UnmarshalWithBSONUnmarshaler(byteSlice []byte, v interface{}) error {
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
