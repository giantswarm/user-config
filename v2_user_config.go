package userconfig

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/giantswarm/generic-types-go"
)

type V2AppDefinition struct {
	Nodes NodeDefinitions `json:"nodes"`
}

func (ad *V2AppDefinition) UnmarshalJSON(data []byte) error {
	// We fix the json buffer so V2CheckForUnknownFields doesn't complain about
	// `Nodes` (with uper N).
	data, err := FixJSONFieldNames(data)
	if err != nil {
		return err
	}

	if err := V2CheckForUnknownFields(data, ad); err != nil {
		return err
	}

	// Just unmarshal the given bytes into the app def struct, since there
	// were no errors.
	var adc v2AppDefCopy
	if err := json.Unmarshal(data, &adc); err != nil {
		return mask(err)
	}

	result := V2AppDefinition(adc)

	// validate app definition without validation context. validation context is
	// given on server side to additionally validate specific definitions.
	if err := result.Validate(nil); err != nil {
		return mask(err)
	}

	*ad = result

	return nil
}

type ValidationContext struct {
	Org       string
	Protocols []string

	MinScaleSize int
	MaxScaleSize int

	MinVolumeSize VolumeSize
	MaxVolumeSize VolumeSize

	PublicDockerRegistry  string
	PrivateDockerRegistry string
}

// validate performs semantic validations of this V2AppDefinition.
// Return the first possible error.
func (ad *V2AppDefinition) Validate(valCtx *ValidationContext) error {
	if len(ad.Nodes) == 0 {
		return maskf(InvalidAppDefinitionError, "nodes must not be empty")
	}

	if err := ad.Nodes.validate(valCtx); err != nil {
		return mask(err)
	}

	return nil
}

// HideDefaults uses the given validation context to determine what definition
// details should be hidden. The caller can clean the definition that way to
// not confuse the user with information he has not set by himself.
func (ad *V2AppDefinition) HideDefaults(valCtx *ValidationContext) (*V2AppDefinition, error) {
	if valCtx == nil {
		return &V2AppDefinition{}, maskf(MissingValidationContextError, "cannot hide defaults")
	}

	ad.Nodes = ad.Nodes.hideDefaults(valCtx)
	return ad, nil
}

// SetDefaults sets necessary default values if not given by the user.
func (ad *V2AppDefinition) SetDefaults(valCtx *ValidationContext) error {
	if valCtx == nil {
		return maskf(MissingValidationContextError, "cannot set defaults")
	}

	ad.Nodes.setDefaults(valCtx)
	return nil
}

type NodeDefinitions map[NodeName]*NodeDefinition

func (nds NodeDefinitions) validate(valCtx *ValidationContext) error {
	for nodeName, node := range nds {
		if err := nodeName.Validate(); err != nil {
			return mask(err)
		}

		// because of defaulting when validating we need to reference the to the
		// address of the node. so its changes effect the app definition after
		// parsing.
		if err := nds[nodeName].validate(valCtx); err != nil {
			return mask(err)
		}

		// detect invalid links
		for _, link := range node.Links {
			nodeFound := false

			for nn, n := range nds {
				if link.Name == nn.String() {
					nodeFound = true

					if !n.Ports.contains(link.Port) {
						return maskf(InvalidNodeDefinitionError, "invalid link to node '%s': does not export port '%s'", nodeName, link.Port)
					}
				}
			}

			if !nodeFound {
				return maskf(InvalidNodeDefinitionError, "invalid link to node '%s': does not exists", link.Name)
			}
		}
	}

	if err := nds.validatePods(); err != nil {
		return Mask(err)
	}

	return nil
}

func (nds NodeDefinitions) hideDefaults(valCtx *ValidationContext) NodeDefinitions {
	for nodeName, node := range nds {
		nds[nodeName] = node.hideDefaults(valCtx)
	}

	return nds
}

func (nds NodeDefinitions) setDefaults(valCtx *ValidationContext) {
	for nodeName, _ := range nds {
		nds[nodeName].setDefaults(valCtx)
	}
}

func (nds *NodeDefinitions) FindByName(name string) (*NodeDefinition, error) {
	for nodeName, nodeDef := range *nds {
		if name == nodeName.String() {
			return nodeDef, nil
		}
	}

	return nil, maskf(NodeNotFoundError, name)
}

// FilterNodes returns a set of all my nodes for which the given predicate returns true.
func (nds *NodeDefinitions) FilterNodes(predicate func(nodeName NodeName, nodeDef *NodeDefinition) bool) NodeDefinitions {
	list := make(NodeDefinitions)
	for nodeName, nodeDef := range *nds {
		if predicate(nodeName, nodeDef) {
			list[nodeName] = nodeDef
		}
	}
	return list
}

// ChildNodes returns a map of all nodes that are a direct child of a node with
// the given name.
func (nds *NodeDefinitions) ChildNodes(name string) NodeDefinitions {
	return nds.FilterNodes(func(nodeName NodeName, nodeDef *NodeDefinition) bool {
		return isDirectChildOf(name, nodeName.String())
	})
}

// ChildNodesRecursive returns a list of all nodes that are a direct child of a node with
// the given name and all child nodes of this children (recursive).
func (nds *NodeDefinitions) ChildNodesRecursive(name string) NodeDefinitions {
	return nds.FilterNodes(func(nodeName NodeName, nodeDef *NodeDefinition) bool {
		return isChildOf(name, nodeName.String())
	})
}

// isDirectChildOf returns true if the given child name is a direct child of the given parent name.
// E.g.
// - isDirectChildOf("a", "a/b") -> true
// - isDirectChildOf("a", "a/b/c") -> false
func isDirectChildOf(parentName, childName string) bool {
	prefix := parentName + "/"
	if !strings.HasPrefix(childName, prefix) {
		return false
	}
	name := childName[len(prefix):]
	if strings.Contains(name, "/") {
		// Grand child
		return false
	}
	return true
}

// isChildOf returns true if the given child name is a child (recursive) of the given parent name.
// E.g.
// - isChildOf("a", "a/b") -> true
// - isChildOf("a", "a/b/c") -> true
func isChildOf(parentName, childName string) bool {
	prefix := parentName + "/"
	if !strings.HasPrefix(childName, prefix) {
		return false
	}
	return true
}

// PodNodes returns a map of all nodes that are part of the pod specified by a node with
// the given name.
func (nds *NodeDefinitions) PodNodes(name string) (NodeDefinitions, error) {
	parent, err := nds.FindByName(name)
	if err != nil {
		return nil, Mask(err)
	}
	if parent.Pod == PodChildren {
		// Collect all direct child nodes that do not have pod set to 'none'.
		return nds.FilterNodes(func(nodeName NodeName, nodeDef *NodeDefinition) bool {
			return isDirectChildOf(name, nodeName.String()) && nodeDef.Pod != PodNone
		}), nil
	} else if parent.Pod == PodInherit {
		// Collect all child nodes that do not have pod set to 'none'.
		noneNames := []NodeName{}
		children := nds.FilterNodes(func(nodeName NodeName, nodeDef *NodeDefinition) bool {
			if !isChildOf(name, nodeName.String()) {
				return false
			}
			if nodeDef.Pod == PodNone {
				noneNames = append(noneNames, nodeName)
				return false
			}
			return true
		})
		// We now  go over the list and remove all children that have some parent with pod='none'
		for _, nodeName := range noneNames {
			for childName, _ := range children {
				if isChildOf(nodeName.String(), childName.String()) {
					// Child of pod='none', remove from list
					delete(children, childName)
				}
			}
		}
		return children, nil
	} else {
		return nil, Mask(errgo.WithCausef(nil, InvalidArgumentError, "Node '%s' a has no pod setting", name))
	}
}

// MountPoints returns a list of all mount points of a node, that is given by
// name
func (nds *NodeDefinitions) MountPoints(name string) ([]string, error) {
	visited := make(map[string]string)
	return nds.mountPointsRecursive(name, visited)
}

// mountPointsRecursive creates a list of all mount points of a node
func (nds *NodeDefinitions) mountPointsRecursive(name string, visited map[string]string) ([]string, error) {
	// prevent cycles
	if _, ok := visited[name]; ok {
		return nil, maskf(InvalidVolumeConfigError, "volume cycle detected in '%s'", name)
	}
	visited[name] = name

	node, err := nds.FindByName(name)
	if err != nil {
		return nil, mask(err)
	}

	// get all mountpoints
	mountPoints := []string{}
	for _, vol := range node.Volumes {
		if vol.Path != "" {
			mountPoints = append(mountPoints, normalizeFolder(vol.Path))
		} else if vol.VolumePath != "" {
			mountPoints = append(mountPoints, normalizeFolder(vol.VolumePath))
		} else if vol.VolumesFrom != "" {
			p, err := nds.mountPointsRecursive(vol.VolumesFrom, visited)
			if err != nil {
				return nil, err
			}
			mountPoints = append(mountPoints, p...)
		}
	}
	return mountPoints, nil
}

// NodeDefinition represents either a runnable service inside a container or a
// node configuration
type NodeDefinition struct {
	// Name of a docker image to use when running a container. The image includes
	// tags. E.g. dockerfile/redis:latest.
	Image ImageDefinition `json:"image" description:"Name of a docker image to use when running a container. The image includes tags."`

	// If given, overwrite the entrypoint of the docker image.
	EntryPoint string `json:"entrypoint,omitempty" description:"If given, overwrite the entrypoint of the docker image."`

	// List of ports a service exposes. E.g. 6379/tcp
	Ports PortDefinitions `json:"ports,omitempty" description:"List of ports this service exposes."`

	// Docker env to inject into docker containers.
	Env EnvList `json:"env,omitempty" description:"List of environment variables used by this service."`

	// Docker volumes to inject into docker containers.
	Volumes VolumeDefinitions `json:"volumes,omitempty" description:"List of volumes to attach to this service."`

	// Arguments for processes inside docker containers.
	Args []string `json:"args,omitempty" description:"List of arguments passed to the entry point of this service."`

	// Domains to bind the port to:  domainName => port, e.g. "www.heise.de" => "80"
	Domains DomainDefinitions `json:"domains,omitempty" description:"List of domains to bind exposed ports to."`

	// Service names required by a service.
	Links LinkDefinitions `json:"links,omitempty" description:"List of dependencies of this service."`

	Expose []ExposeDefinition `json:"expose,omitempty" description:"List of port mappings to define a stable API."`

	Scale *ScaleDefinition `json:"scale,omitempty" description:"Scaling settings of the node."`

	Pod PodEnum `json:"pod,omitempty" description:"Pod behavior of this node and its children."`
}

// validate performs semantic validations of this NodeDefinition.
// Return the first possible error.
func (nd *NodeDefinition) validate(valCtx *ValidationContext) error {
	if err := nd.Image.Validate(valCtx); err != nil {
		return mask(err)
	}

	if err := nd.Ports.Validate(valCtx); err != nil {
		return mask(err)
	}

	if err := nd.Domains.validate(nd.Ports); err != nil {
		return mask(err)
	}

	if err := nd.Links.Validate(valCtx); err != nil {
		return mask(err)
	}

	if err := nd.Volumes.validate(valCtx); err != nil {
		return mask(err)
	}

	if nd.Scale != nil {
		if err := nd.Scale.validate(valCtx); err != nil {
			return mask(err)
		}
	}

	return nil
}

func (nd *NodeDefinition) hideDefaults(valCtx *ValidationContext) *NodeDefinition {
	if nd.Scale != nil {
		nd.Scale = nd.Scale.hideDefaults(valCtx)
	}

	return nd
}

func (nd *NodeDefinition) setDefaults(valCtx *ValidationContext) {
	// set default scale definition if not set
	if nd.Scale == nil {
		nd.Scale = &ScaleDefinition{}
	}

	nd.Scale.setDefaults(valCtx)
}

type ExposeDefinition struct {
	Port     generictypes.DockerPort `json:"port" description:"Port of the stable API."`
	Node     string                  `json:"node" description:"Node name of the node that exposes a given port."`
	NodePort generictypes.DockerPort `json:"node_port" description:"Port of the given node."`
}

// TODO Node.IsService() bool

// V2GenerateAppName removes any formatting from b and returns the first 4 bytes
// of its MD5 checksum.
func V2GenerateAppName(b []byte) (string, error) {
	// parse and validate
	appDef, err := ParseV2AppDefinition(b)
	if err != nil {
		return "", mask(err)
	}

	// remove formatting
	clean, err := json.Marshal(appDef)
	if err != nil {
		return "", mask(err)
	}

	// create hash
	s := md5.Sum(clean)
	return fmt.Sprintf("%x", s[0:4]), nil
}

// ParseV2AppDefinition tries to parse the v2 app definition.
func ParseV2AppDefinition(b []byte) (V2AppDefinition, error) {
	var appDef V2AppDefinition
	if err := json.Unmarshal(b, &appDef); err != nil {
		if IsSyntax(err) {
			if strings.Contains(err.Error(), "$") {
				return V2AppDefinition{}, maskf(err, "Cannot parse swarm.json. Maybe not all variables replaced properly.")
			}
		}

		return V2AppDefinition{}, mask(err)
	}

	return appDef, nil
}
