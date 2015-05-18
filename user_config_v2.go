package userconfig

import (
	"encoding/json"
	"reflect"
	"strings"

	"github.com/giantswarm/docker-types-go"
	"github.com/juju/errgo"
)

const (
	pathSeparator = "/"
)

//
type NodeName struct {
	content string
	root    string
	dir     string
	base    string
}

func (nn *NodeName) parse(input string) *NodeName {
	splitted := strings.Split(input, pathSeparator)

	// In case the node name starts with a slash, the first element is empty.
	if splitted[0] == "" {
		splitted = splitted[1:]
	}

	// In case the node name ends with a slash, the last element is empty.
	lastIndex := len(splitted) - 1
	if splitted[lastIndex] == "" {
		splitted = splitted[:lastIndex]
	}

	newnn := &NodeName{
		content: strings.Join(splitted, pathSeparator),
		root:    splitted[0],
		dir:     splitted[0],
		base:    splitted[0],
	}

	if len(splitted) > 1 {
		lastIndex := len(splitted) - 1

		newnn.dir = strings.Join(splitted[:lastIndex], pathSeparator)
		newnn.base = strings.Join(splitted[lastIndex:], pathSeparator)
	}

	return newnn
}

func (nn *NodeName) UnmarshalJSON(data []byte) error {
	var input string

	if err := json.Unmarshal(data, &input); err != nil {
		return errgo.Mask(err)
	}

	*nn = *nn.parse(input)

	return nil
}

func (nn *NodeName) MarshalJSON() ([]byte, error) {
	return json.Marshal(nn.String())
}

func (nn *NodeName) String() string {
	return string(nn.content)
}

// Root returns the first element of the node name separated by a file path
// separator, e.g. "/".
func (nn *NodeName) Root() string {
	return nn.root
}

// Dir returns the directory, in which Base is located.
func (nn *NodeName) Dir() string {
	return nn.dir
}

// Base returns the last element of the node name separated by a file path
// separator, e.g. "/".
func (nn *NodeName) Base() string {
	return nn.base
}

type LinkConf struct {
	Node  *NodeName              `json:"node,omitempty"`
	Alias string                 `json:"alias,omitempty"`
	Port  dockertypes.DockerPort `json:"port,omitempty"`
}

type Node struct {
	Name    *NodeName                `json:"name,omitempty"`
	Ports   []dockertypes.DockerPort `json:"ports,omitempty"`
	Image   dockertypes.DockerImage  `json:"image,omitempty"`
	Scale   ScalingPolicyConfig      `json:"scale,omitempty"`
	Domains DomainConfig             `json:"domains,omitempty"`
	Links   []LinkConf               `json:"links,omitempty"`
	Env     EnvList                  `json:"env,omitempty"`
	Volumes []VolumeConfig           `json:"volumes,omitempty"`
	Args    []string                 `json:"args,omitempty"`
}

func (n Node) IsScalingNode() bool {
	ref := Node{Name: n.Name, Scale: n.Scale}
	return reflect.DeepEqual(ref, n)
}

type V2AppDefinition []Node

func (ad V2AppDefinition) AppName() string {
	for _, node := range ad {
		if !node.IsScalingNode() {
			return node.Name.Root()
		}
	}

	// this should never happen
	panic("cannot find app name")

	return ""
}
