package userconfig

import (
	"encoding/json"
	"fmt"
	"sort"

	"github.com/giantswarm/generic-types-go"
)

type V2DomainDefinitions map[generictypes.Domain]PortDefinitions
type domainList []generictypes.Domain

// UnmarshalJSON performs custom unmarshalling to support smart
// data types.
func (dds *V2DomainDefinitions) UnmarshalJSON(data []byte) error {
	// Try format 1: domain: port
	err := dds.unmarshalJSONDomainPortMap(data)
	if err == nil {
		// Found a correct result
		return nil
	}

	// Try format 2: port: domainList
	if err2 := dds.unmarshalJSONPortDomainList(data); err2 == nil {
		// Found a correct result
		return nil
	}

	// Unknown format
	return maskf(InvalidDomainDefinitionError, "invalid format for domains: %s", err.Error())
}

// MarshalJSON performs custom marshalling to generate the reverse format:
// port: domainList
func (dds V2DomainDefinitions) MarshalJSON() ([]byte, error) {
	portDomainMap := make(map[string]domainList)
	for domain, ports := range dds {
		for _, port := range ports {
			portStr := port.String()
			list, ok := portDomainMap[portStr]
			if !ok {
				list = domainList{}
			}
			list = append(list, domain)
			portDomainMap[portStr] = list
		}
	}

	data, err := json.Marshal(portDomainMap)
	if err != nil {
		return nil, mask(err)
	}
	return data, nil
}

// unmarshalJSONDomainPortMap tries to unmarshal the given data
// in format: domain: port, domain2: port2
func (dds *V2DomainDefinitions) unmarshalJSONDomainPortMap(data []byte) error {
	var local map[generictypes.Domain]PortDefinitions
	if err := json.Unmarshal(data, &local); err == nil {
		// Found a correct result
		*dds = V2DomainDefinitions(local)
		return nil
	} else {
		return mask(err)
	}
}

// unmarshalJSONPortDomainList tries to unmarshal the given data
// in format: port: domainList, port2: domainList...
func (dds *V2DomainDefinitions) unmarshalJSONPortDomainList(data []byte) error {
	var local map[string]domainList
	if err := json.Unmarshal(data, &local); err == nil {
		// Found a correct result, convert it
		newMap := V2DomainDefinitions{}
		for p, list := range local {
			port, err := generictypes.ParseDockerPort(p)
			if err != nil {
				return mask(err)
			}
			for _, domain := range list {
				newMap[domain] = PortDefinitions{port}
			}
		}
		*dds = newMap
		return nil
	} else {
		return mask(err)
	}
}

// UnmarshalJSON performs custom unmarshalling to support smart
// data types.
func (dl *domainList) UnmarshalJSON(data []byte) error {
	if data[0] != '[' {
		// Must be a single value, convert to an array of one
		newData := []byte{}
		newData = append(newData, '[')
		newData = append(newData, data...)
		newData = append(newData, ']')

		data = newData
	}

	var local []generictypes.Domain
	if err := json.Unmarshal(data, &local); err != nil {
		return mask(err)
	}
	*dl = domainList(local)

	return nil
}

// String returns the marshalled and ordered string represantion of its own
// incarnation. It is important to have the string represantion ordered, since
// we use it to compare two V2DomainDefinitions when creating a diff. See diff.go
func (dds V2DomainDefinitions) String() string {
	keys := []string{}
	for domain, _ := range dds {
		keys = append(keys, domain.String())
	}
	sort.Strings(keys)

	simple := map[string]string{}
	for _, key := range keys {
		ports := dds[generictypes.Domain(key)]
		simple[key] = ports.String()
	}

	raw, err := json.Marshal(simple)
	if err != nil {
		panic(fmt.Sprintf("%#v\n", mask(err)))
	}

	return string(raw)
}

func (dds V2DomainDefinitions) validate(exportedPorts PortDefinitions) error {
	for domainName, ports := range dds {
		if err := domainName.Validate(); err != nil {
			return mask(err)
		}

		for _, port := range ports {
			if !exportedPorts.contains(port) {
				return maskf(InvalidDomainDefinitionError, "port '%s' of domain '%s' must be exported", port, domainName)
			}
		}
	}

	return nil
}
