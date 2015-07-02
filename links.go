package userconfig

type LinkDefinitions []DependencyConfig

func (lds LinkDefinitions) Validate(valCtx *ValidationContext) error {
	links := map[string]bool{}

	for _, link := range lds {
		if err := link.Validate(valCtx); err != nil {
			return mask(err)
		}

		// detect duplicated link name
		name := link.Alias
		if name == "" {
			name = link.Name
		}
		if _, ok := links[name]; ok {
			return maskf(InvalidLinkDefinitionError, "duplicated link: %s", name)
		}
		links[name] = true
	}

	return nil
}
