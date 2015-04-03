package userconfig

import (
  "reflect"

	"github.com/giantswarm/docker-types-go"
)

type NodeName string

func (nn NodeName) Base() string {
  return filepath.Base(nn)
}

func (nn NodeName) Dir() string {
  return filepath.Dir(nn)
}

type LinkConf struct {
  Node NodeName   `json:"node,omitempty"`
	Port dockertypes.DockerPort `json:"port,omitempty"`
}

type Node struct {
  Name    NodeName   `json:"name,omitempty"`
	Ports []dockertypes.DockerPort `json:"ports,omitempty"`
	Image dockertypes.DockerImage `json:"image,omitempty"`
  Scale   ScalingPolicyConfig      `json:"scale,omitempty"`
  Domains DomainConf `json:"domains,omitempty"`
  Links   []LinkConf `json:"links,omitempty"`
	Env EnvList `json:"env,omitempty"`
	Volumes []VolumeConfig `json:"volumes,omitempty"`
	Args []string `json:"args,omitempty"`
}

func (n Node) IsScalingNode() bool {
  ref := Node{Node: n.Node, Scale: n.Scale}
  return reflect.DeepEqual(ref, n)
}

type AppDefinitionV2 []Node
