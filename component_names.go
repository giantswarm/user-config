package userconfig

import (
	"encoding/json"
	"fmt"
)

type ComponentNames []ComponentName

func (cns ComponentNames) Contain(name ComponentName) bool {
	for _, cn := range cns {
		if cn.Equals(name) {
			return true
		}
	}

	return false
}

// NamesToJSONString returns a JSON marshaled string of component names.
func (cns ComponentNames) ToJSONString() string {
	raw, err := json.Marshal(cns)
	if err != nil {
		panic(fmt.Sprintf("%#v", maskAny(err)))
	}

	return string(raw)
}
