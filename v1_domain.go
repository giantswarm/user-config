package userconfig

import (
	"github.com/giantswarm/generic-types-go"
)

type DomainDefinitions map[generictypes.Domain]generictypes.DockerPort

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
