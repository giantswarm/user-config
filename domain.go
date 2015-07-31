package userconfig

import (
	"encoding/json"

	"github.com/giantswarm/generic-types-go"
)

type DomainDefinitions map[generictypes.Domain]generictypes.DockerPort
type domainList []generictypes.Domain

// UnmarshalJSON performs custom unmarshalling to support smart
// data types.
func (dds *DomainDefinitions) UnmarshalJSON(data []byte) error {
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

// unmarshalJSONDomainPortMap tries to unmarshal the given data
// in format: domain: port, domain2: port2
func (dds *DomainDefinitions) unmarshalJSONDomainPortMap(data []byte) error {
	var local map[generictypes.Domain]generictypes.DockerPort
	if err := json.Unmarshal(data, &local); err == nil {
		// Found a correct result
		*dds = DomainDefinitions(local)
		return nil
	} else {
		return mask(err)
	}
}

// unmarshalJSONPortDomainList tries to unmarshal the given data
// in format: port: domainList, port2: domainList...
func (dds *DomainDefinitions) unmarshalJSONPortDomainList(data []byte) error {
	var local map[string]domainList
	if err := json.Unmarshal(data, &local); err == nil {
		// Found a correct result, convert it
		newMap := DomainDefinitions{}
		for p, list := range local {
			port, err := generictypes.ParseDockerPort(p)
			if err != nil {
				return mask(err)
			}
			for _, domain := range list {
				newMap[domain] = port
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

// ToSimple just maps the domain config with its custom types to a more simpler
// map. This can be used for internal management once the validity of domains
// and ports is given. That way dependencies between packages requiring hard
// custom types can be dropped.
func (dc DomainDefinitions) ToSimple() map[string]string {
	simpleDomains := map[string]string{}

	for d, p := range dc {
		simpleDomains[d.String()] = p.Port
	}

	return simpleDomains
}

func (dc DomainDefinitions) validate(exportedPorts PortDefinitions) error {
	for domainName, domainPort := range dc {
		if err := domainName.Validate(); err != nil {
			return mask(err)
		}

		if !exportedPorts.contains(domainPort) {
			return maskf(InvalidDomainDefinitionError, "port '%s' of domain '%s' must be exported", domainPort.Port, domainName)
		}
	}

	return nil
}
