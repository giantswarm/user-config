package userconfig

import (
	"github.com/giantswarm/generic-types-go"
	"github.com/juju/errgo"
)

type PortDefinitions []generictypes.DockerPort

func (pds PortDefinitions) validate() error {
	for _, port := range pds {
		if port.Protocol != generictypes.ProtocolTCP {
			return Mask(errgo.WithCausef(nil, InvalidPortConfigError, "invalid protocol '%s' for port '%s', expected %s", port.Protocol, port.Port, generictypes.ProtocolTCP))
		}
	}

	return nil
}

func (pds PortDefinitions) contains(port generictypes.DockerPort) bool {
	for _, pd := range pds {
		// generictypes.DockerPort implements Equals to properly compare the
		// format "<port>/<protocol>"
		if pd.Equals(port) {
			return true
		}
	}

	return false
}
