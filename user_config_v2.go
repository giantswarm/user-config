package userconfig

import (
	"encoding/json"
	"reflect"

	"github.com/dchest/uniuri"
	"github.com/giantswarm/docker-types-go"
)

type NodeName string

func (nn NodeName) UnmarshalJSON(data []byte) error {
	var input string

	if err := json.Unmarshal(data, &input); err != nil {
		return errgo.Mask(err)
	}

	return input
}

func (nn NodeName) MarshalJSON() ([]byte, error) {
	return json.Marshal(nn.String())
}

func (nn NodeName) String() string {
	return nn
}

func (nn NodeName) Base() string {
	return filepath.Base(nn.String())
}

func (nn NodeName) Root() string {
	return strings.Split(nn.String(), filepath.Separator)[0]
}

func (nn NodeName) Dir() string {
	return filepath.Dir(nn.String())
}

type LinkConf struct {
	Node  NodeName               `json:"node,omitempty"`
	Alias string                 `json:"alias,omitempty"`
	Port  dockertypes.DockerPort `json:"port,omitempty"`
}

type Node struct {
	Name    NodeName                 `json:"name,omitempty"`
	Ports   []dockertypes.DockerPort `json:"ports,omitempty"`
	Image   dockertypes.DockerImage  `json:"image,omitempty"`
	Scale   ScalingPolicyConfig      `json:"scale,omitempty"`
	Domains DomainConf               `json:"domains,omitempty"`
	Links   []LinkConf               `json:"links,omitempty"`
	Env     EnvList                  `json:"env,omitempty"`
	Volumes []VolumeConfig           `json:"volumes,omitempty"`
	Args    []string                 `json:"args,omitempty"`
}

func (n Node) IsScalingNode() bool {
	ref := Node{Node: n.Node, Scale: n.Scale}
	return reflect.DeepEqual(ref, n)
}

type AppDefinitionV2 []Node
