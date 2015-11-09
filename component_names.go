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

func (cns ComponentNames) Filter(names ComponentNames) ComponentNames {
	list := ComponentNames{}

	for _, cn := range cns {
		if names.Contain(cn) {
			continue
		}
		list = append(list, cn)
	}

	return list
}

func (cns ComponentNames) ContainAny(names ComponentNames) bool {
	for _, cn := range cns {
		for _, name := range names {
			if cn.Equals(name) {
				return true
			}
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
