package userconfig

import (
	"encoding/json"

	"github.com/giantswarm/docker-types-go"
	"github.com/alexanderritola/validate"
	"github.com/alexanderritola/validate/web"
	"github.com/juju/errgo"
)

type Domain string

type DomainConfig map[Domain]dockertypes.DockerPort

func (dc DomainConfig) ToSimple() map[string]string {
  simpleDomains := map[string]string{}

	for d, p := range dc {
		simpleDomains[d.String()] = p.Port
	}

	return simpleDomains
}

func (d *Domain) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.String())
}

func (d *Domain) UnmarshalJSON(data []byte) error {
	var input string

	if err := json.Unmarshal(data, &input); err != nil {
		return err
	}

	*d = Domain(input)

	if err := d.Validate(); err != nil {
		return err
	}

	return nil
}

func (d *Domain) String() string {
	return string(*d)
}

func (d *Domain) Validate() error {
	v := validate.NewValidator()

	if err := v.Validate(web.NewDomain(d.String())); err != nil {
		return errgo.Mask(errgo.Newf("Invalid domain: %s", d.String()), errgo.Any)
	}

	return nil
}
