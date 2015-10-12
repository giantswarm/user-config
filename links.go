// TODO this file should be called link.go
package userconfig

import (
	"encoding/json"
	"fmt"
	"sort"

	"github.com/giantswarm/generic-types-go"
)

type LinkDefinition struct {
	Service AppName `json:"service,omitempty" description:"Name of the service that is linked to"`

	// Name of a required component
	Component ComponentName `json:"component,omitempty" description:"Name of a component that is linked to"`

	// The name how this dependency should appear in the container
	Alias string `json:"alias,omitempty" description:"The name how this dependency should appear in the container"`

	// Port of the required component
	TargetPort generictypes.DockerPort `json:"target_port" description:"Port on the component that is linked to"`
}

type LinkDefinitions []LinkDefinition

// String returns the string represantion of the current incarnation.
func (ld LinkDefinition) String() string {
	// A string map is reliable enough for our case, as the JSON implementation
	// takes care of the order of the provided fields. See
	// http://play.golang.org/p/U8nDgdga2X
	m := map[string]string{
		"service":     ld.Service.String(),
		"component":   ld.Component.String(),
		"alias":       ld.Alias,
		"target_port": ld.TargetPort.String(),
	}

	raw, err := json.Marshal(m)
	if err != nil {
		panic(fmt.Sprintf("%#v\n", mask(err)))
	}

	return string(raw)
}

func (ld LinkDefinition) Validate(valCtx *ValidationContext) error {
	if ld.Component.Empty() && ld.Service.Empty() {
		return maskf(InvalidLinkDefinitionError, "link component must not be empty")
	}
	if !ld.Component.Empty() {
		if err := ld.Component.Validate(); err != nil {
			return maskf(InvalidLinkDefinitionError, "invalid link component: %s", err.Error())
		}
	}
	if !ld.Service.Empty() {
		if err := ld.Service.Validate(); err != nil {
			return maskf(InvalidLinkDefinitionError, "invalid link service: %s", err.Error())
		}
	}
	if !ld.Component.Empty() && !ld.Service.Empty() {
		return maskf(InvalidLinkDefinitionError, "link service and component cannot be set both")
	}

	// for easy validation we create a port definitions type and use its
	// validate method
	pds := PortDefinitions{ld.TargetPort}
	if err := pds.Validate(valCtx); err != nil {
		return maskf(InvalidLinkDefinitionError, "invalid link: %s", err.Error())
	}

	return nil
}

// LinkName returns the name of this link as it will be used inside
// the component.
// This defaults to the alias. If that is not specified, the local name
// of the Component name will be used, or if that is also empty, the app name.
func (ld LinkDefinition) LinkName() (string, error) {
	if ld.Alias != "" {
		return ld.Alias, nil
	}
	if !ld.Component.Empty() {
		// Take the dependency name from the last part of the component name
		// (using `LocalName()`).
		// This is done to prevent that the dependency name has '/' in it.
		return ld.Component.LocalName().String(), nil
	}
	if !ld.Service.Empty() {
		return ld.Service.String(), nil
	}
	return "", mask(InvalidLinkDefinitionError)
}

// LinksToOtherService returns true if this definition defines
// a link between a component an another service.
func (ld LinkDefinition) LinksToOtherService() bool {
	return !ld.Service.Empty()
}

// LinksToSameService returns true if this definition defines
// a link between a component and another component within the same service.
func (ld LinkDefinition) LinksToSameService() bool {
	return ld.Service.Empty()
}

func (lds LinkDefinitions) Validate(valCtx *ValidationContext) error {
	links := map[string]string{}

	for _, link := range lds {
		if err := link.Validate(valCtx); err != nil {
			return mask(err)
		}

		// detect duplicated link name
		linkName, err := link.LinkName()
		if err != nil {
			return mask(err)
		}
		if _, ok := links[linkName]; ok {
			return maskf(InvalidLinkDefinitionError, "duplicate link: %s", linkName)
		}
		links[linkName] = link.TargetPort.String()
	}

	return nil
}

// String returns the marshalled and ordered string represantion of its own
// incarnation. It is important to have the string represantion ordered, since
// we use it to compare two LinkDefinitions when creating a diff. See diff.go
func (lds LinkDefinitions) String() string {
	list := []string{}

	for _, ld := range lds {
		list = append(list, ld.String())
	}
	sort.Strings(list)

	raw, err := json.Marshal(list)
	if err != nil {
		panic(fmt.Sprintf("%#v\n", mask(err)))
	}

	return string(raw)
}

// Resolve resolves the implementation of the given link in the context of the given
// component definitions.
// Resolve returns the name of the component that implements this link and its implementation port.
// If this link cannot be resolved, an error is returned.
func (link LinkDefinition) Resolve(nds ComponentDefinitions) (ComponentName, generictypes.DockerPort, error) {
	// Resolve initial link target
	targetName := link.Component
	targetComponent, err := nds.ComponentByName(targetName)
	if err != nil {
		return "", generictypes.DockerPort{}, maskf(ComponentNotFoundError, link.Component.String())
	}

	// If the linked to port exposed by the target component?
	if expDef, err := targetComponent.Expose.defByPort(link.TargetPort); err == nil {
		// Link to exposed port, let expose definition resolve this further
		return expDef.Resolve(targetName, nds)
	}

	if targetComponent.Ports.contains(link.TargetPort) {
		// Link points directly to an exported port of the target
		return targetName, link.TargetPort, nil
	}

	// Invalid link
	return "", generictypes.DockerPort{}, maskf(InvalidLinkDefinitionError, "port %s not found in %s", link.TargetPort, targetName)
}

// validateLinks
func (nds ComponentDefinitions) validateLinks() error {
	for componentName, component := range nds {
		// detect invalid links
		for _, link := range component.Links {
			// If the link is inter-service, we cannot validate it here.
			if link.LinksToOtherService() {
				continue
			}

			// Try to find the target component
			targetName := ComponentName(link.Component)
			targetComponent, err := nds.ComponentByName(targetName)
			if IsComponentNotFound(err) {
				return maskf(InvalidComponentDefinitionError, "invalid link to component '%s': does not exists", link.Component)
			} else if err != nil {
				return maskf(InvalidComponentDefinitionError, "unexpected error: %#v", err)
			}

			// Does the target component expose the linked to port?
			if !targetComponent.Expose.contains(link.TargetPort) && !targetComponent.Ports.contains(link.TargetPort) {
				return maskf(InvalidComponentDefinitionError, "invalid link to component '%s': does not export port '%s'", link.Component, link.TargetPort)
			}

			// Is the component allowed to link to the target component?
			if !isLinkAllowed(componentName, targetName) {
				return maskf(InvalidLinkDefinitionError, "invalid link to component '%s': component '%s' is not allowed to link to it", link.Component, componentName)
			}

			if err := nds.detectLinkCycle(link); err != nil {
				return maskf(InvalidComponentDefinitionError, "invalid link to component '%s': %s", link.Component, err.Error())
			}
		}
	}

	return nil
}

// detectLinkCycle walks the links of the given component, looks up the target
// components of each link, and follows this components links, until it finds a
// loop. In case there is no loop, it does not throw an error. In case it
// detects a loop, it throws an error.
func (nds ComponentDefinitions) detectLinkCycle(linkDefinition LinkDefinition) error {
	linkedComponents := ComponentNames{}

	var recursive func(ld LinkDefinition) error
	recursive = func(ld LinkDefinition) error {
		targetName := ComponentName(ld.Component)

		if linkedComponents.Contain(targetName) {
			// We found a loop.
			return maskAny(LinkCycleError)
		}
		linkedComponents = append(linkedComponents, targetName)

		targetComponent, err := nds.ComponentByName(targetName)
		if err != nil {
			return maskAny(err)
		}

		// Go deeper into the dependency graph
		for _, tcl := range targetComponent.Links {
			if err := recursive(tcl); err != nil {
				return maskAny(err)
			}
		}

		return nil
	}

	if err := recursive(linkDefinition); err != nil {
		return maskAny(err)
	}

	return nil
}

// isLinkAllowed returns true if a component with given name is allowed to
// link to a component with given target name.
func isLinkAllowed(componentName, targetName ComponentName) bool {
	// If target is a child or grand child of component, it is ok.
	if targetName.IsChildOf(componentName) {
		return true
	}

	// If target is a parent/sibling ("up or right-left"), it is ok.
	if isParentOrSiblingRecursive(componentName, targetName) {
		return true
	}

	return false
}

// isParentOrSiblingRecursive returns true if targetName is a parent of componentName,
// or targetName is a sibling of component name.
// The test is done recursively.
func isParentOrSiblingRecursive(componentName, targetName ComponentName) bool {
	if componentName.IsSiblingOf(targetName) {
		return true
	}
	parentName, err := componentName.ParentName()
	if err != nil {
		// No more parent
		return false
	}
	return isParentOrSiblingRecursive(parentName, targetName)
}
