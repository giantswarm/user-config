package userconfig

import (
//"encoding/json"
//"fmt"

//"github.com/juju/errgo"
)

func CheckForUnknownFields(b []byte, ac *AppConfig) error {
	// Unmarshal clean struct
	//var clean AppConfig
	//if err := json.Unmarshal(b, &clean); err != nil {
	//	return Mask(err)
	//}

	//// Marshal clean struct
	//cleanBytes, err := json.Marshal(clean)
	//if err != nil {
	//	return Mask(err)
	//}

	//fmt.Printf("%#v\n", string(cleanBytes))
	//fmt.Printf("%#v\n", string(b))

	// compare bytes
	// if diverged unmarshal into map[string]interface{}
	//   loop and detect diff

	//		// Better reset app-config to its zero value.
	//		*appConfig = AppConfig{}

	//		return errgo.WithCausef(nil, ErrUnknownJSONField, "Cannot parse app config. Unknown field '%s' detected.", k)

	return nil
}

//func UnmarshalDirty(byteSlice []byte, v interface{}) error {
//	var j map[string]interface{}
//
//	if err := json.Unmarshal(byteSlice, &j); err != nil {
//		return Mask(err)
//	}
//
//	byteSlice, err := json.Marshal(&j)
//	if err != nil {
//		return Mask(err)
//	}
//
//	if err := json.Unmarshal(byteSlice, v); err != nil {
//		return Mask(err)
//	}
//
//	return nil
//}
