package userconfig

import (
	"encoding/json"
	"fmt"
	"sort"

	"github.com/juju/errgo"
)

// List of environment settings like "KEY=VALUE", "KEY2=VALUE2"
type EnvList []string

// UnmarshalJSON supports parsing an EnvList as array and as structure
func (this *EnvList) UnmarshalJSON(data []byte) error {
	var err error
	// Try to parse as struct first
	if len(data) > 1 && data[0] == '{' {
		kvMap := make(map[string]string)
		err = json.Unmarshal(data, &kvMap)
		if err == nil {
			// Success, wrap into array
			// Sort the keys first so the outcome it always the same
			keys := []string{}
			for k, _ := range kvMap {
				keys = append(keys, k)
			}
			sort.Strings(keys)
			list := []string{}
			for _, k := range keys {
				v := kvMap[k]
				list = append(list, fmt.Sprintf("%s=%s", k, v))
			}
			*this = list
			return nil
		}
	}

	// Try to parse are []string
	if len(data) > 1 && data[0] == '[' {
		list := []string{}
		err = json.Unmarshal(data, &list)
		if err != nil {
			return err
		}
		*this = list
		return nil
	}

	return errgo.WithCausef(err, InvalidEnvListFormatError, "")
}

func (eds *EnvList) String() string {
	list := []string{}

	for _, ed := range *eds {
		list = append(list, ed)
	}
	sort.Strings(list)

	raw, err := json.Marshal(list)
	if err != nil {
		panic(fmt.Sprintf("%#v\n", mask(err)))
	}

	return string(raw)
}
