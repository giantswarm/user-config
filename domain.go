package userconfig

import (
	"github.com/giantswarm/generic-types-go"
)

type DomainConfig map[generictypes.Domain]generictypes.DockerPort

// ToSimple just maps the domain config with its custom types to a more simpler
// map. This can be used for internal management once the validity of domains
// and ports is given. That way dependencies between packages requiring hard
// custom types can be dropped.
func (dc DomainConfig) ToSimple() map[string]string {
	simpleDomains := map[string]string{}

	for d, p := range dc {
		simpleDomains[d.String()] = p.Port
	}

	return simpleDomains
}
