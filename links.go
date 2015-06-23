package userconfig

import (
	"github.com/juju/errgo"
)

type LinkDefinitions []DependencyConfig

func (lds LinkDefinitions) validate() error {
	links := map[string]bool{}

	for _, link := range lds {
		if err := link.validate(); err != nil {
			return Mask(err)
		}

		// detect duplicated link name
		name := link.Alias
		if name == "" {
			name = link.Name
		}
		if _, ok := links[name]; ok {
			return Mask(errgo.WithCausef(nil, InvalidLinkDefinitionError, "duplicated link: %s", name))
		}
		links[name] = true
	}

	return nil
}
