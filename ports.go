package userconfig

import (
	"github.com/giantswarm/generic-types-go"
	"github.com/juju/errgo"
)

type PortDefinitions []generictypes.DockerPort

// Validate tries to validate the current PortDefinitions. If valCtx is nil,
// nothing can be validated. The given valCtx must provide at least one
// protocol, or Validate returns an error. The currently valid one should only
// be TCP.
func (pds PortDefinitions) Validate(valCtx *ValidationContext) error {
	if valCtx == nil {
		return nil
	}

	if len(valCtx.Protocols) == 0 {
		return errgo.Newf("missing protocol in validation context")
	}

	for _, port := range pds {
		if !contains(valCtx.Protocols, port.Protocol) {
			return mask(errgo.WithCausef(nil, InvalidPortConfigError, "invalid protocol '%s' for port '%s', expected one of %v", port.Protocol, port.Port, valCtx.Protocols))
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

func contains(protocols []string, protocol string) bool {
	for _, p := range protocols {
		if p == protocol {
			return true
		}
	}

	return false
}
