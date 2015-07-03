package userconfig

import (
	"encoding/json"
	"testing"

	"github.com/giantswarm/generic-types-go"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestV2UserConfigPodValidator(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "v2 user config pod validator")
}

var _ = Describe("v2 user config pod validator", func() {

	testNode := func() *NodeDefinition {
		return &NodeDefinition{
			Image: MustParseImageDefinition("registry/namespace/repository:version"),
		}
	}

	setPod := func(config *NodeDefinition, pod PodEnum) *NodeDefinition {
		config.Pod = pod
		return config
	}

	addVols := func(config *NodeDefinition, vol ...VolumeConfig) *NodeDefinition {
		if config.Volumes == nil {
			config.Volumes = vol
		} else {
			config.Volumes = append(config.Volumes, vol...)
		}
		return config
	}

	addDeps := func(config *NodeDefinition, dep ...DependencyConfig) *NodeDefinition {
		if config.Links == nil {
			config.Links = dep
		} else {
			config.Links = append(config.Links, dep...)
		}
		return config
	}

	addPorts := func(config *NodeDefinition, ports ...generictypes.DockerPort) *NodeDefinition {
		if config.Ports == nil {
			config.Ports = ports
		} else {
			config.Ports = append(config.Ports, ports...)
		}
		return config
	}

	addScale := func(config *NodeDefinition, min, max int) *NodeDefinition {
		config.Scale = &ScaleDefinition{
			Min: min,
			Max: max,
		}
		return config
	}

	testApp := func() NodeDefinitions {
		return make(NodeDefinitions)
	}

	maxScale := 7
	validate := func(nds NodeDefinitions) error {
		vctx := &ValidationContext{
			Protocols:     []string{"tcp"},
			MinVolumeSize: "1 GB",
			MaxVolumeSize: "100 GB",
			MinScaleSize:  1,
			MaxScaleSize:  maxScale,
		}
		nds.setDefaults(vctx)
		if err := nds.validate(vctx); err != nil {
			return mask(err)
		}
		return nil
	}

	Describe("pod tests ", func() {
		Describe("parsing (valid) pod==children on a node that has 2 children", func() {
			var err error

			BeforeEach(func() {
				nodes := testApp()
				nodes["node/a"] = setPod(testNode(), PodChildren)
				nodes["node/a/b"] = testNode()
				nodes["node/a/c"] = testNode()

				err = validate(nodes)
			})

			It("should not throw any errors", func() {
				Expect(err).To(BeNil())
			})
		})

		Describe("parsing (invalid) pod==children on a node that has no children", func() {
			var err error

			BeforeEach(func() {
				nodes := testApp()
				nodes["node/a"] = setPod(testNode(), PodChildren)

				err = validate(nodes)
			})

			It("should throw error IsInvalidPodConfig", func() {
				Expect(IsInvalidPodConfig(err)).To(BeTrue())
				Expect(err.Error()).To(Equal(`Node 'node/a' must have at least 2 child nodes because if defines 'pod' as 'children'`))
			})
		})

		Describe("parsing (invalid) pod==children on a node that has 2 children (recursive), but only 1 direct", func() {
			var err error

			BeforeEach(func() {
				nodes := testApp()
				nodes["node/a"] = setPod(testNode(), PodChildren)
				nodes["node/a/b"] = testNode()
				nodes["node/a/b/c"] = testNode()

				err = validate(nodes)
			})

			It("should throw error IsInvalidPodConfig", func() {
				Expect(IsInvalidPodConfig(err)).To(BeTrue())
				Expect(err.Error()).To(Equal(`Node 'node/a' must have at least 2 child nodes because if defines 'pod' as 'children'`))
			})
		})

		Describe("parsing (valid) pod==inherit on a node that has 2 children (recursive)", func() {
			var err error

			BeforeEach(func() {
				nodes := testApp()
				nodes["node/a"] = setPod(testNode(), PodInherit)
				nodes["node/a/b"] = testNode()
				nodes["node/a/b/c"] = testNode()

				err = validate(nodes)
			})

			It("should not throw any errors", func() {
				Expect(err).To(BeNil())
			})
		})

		Describe("parsing (invalid) pod==inherit on a node that has not enough children", func() {
			var err error

			BeforeEach(func() {
				nodes := testApp()
				nodes["node/a"] = setPod(testNode(), PodInherit)
				nodes["node/a/b"] = testNode()

				err = validate(nodes)
			})

			It("should throw error IsInvalidPodConfig", func() {
				Expect(IsInvalidPodConfig(err)).To(BeTrue())
				Expect(err.Error()).To(Equal(`Node 'node/a' must have at least 2 child nodes because if defines 'pod' as 'inherit'`))
			})
		})

		Describe("parsing (invalid) cannot specify pod value other than 'none' inside a pod", func() {
			var err error

			BeforeEach(func() {
				nodes := testApp()
				nodes["node/a"] = setPod(testNode(), PodChildren)
				nodes["node/a/b"] = setPod(testNode(), PodChildren)
				nodes["node/a/b/c1"] = testNode()
				nodes["node/a/b/c2"] = testNode()
				nodes["node/a/c"] = testNode()

				err = validate(nodes)
			})

			It("should throw error IsInvalidPodConfig", func() {
				Expect(IsInvalidPodConfig(err)).To(BeTrue())
				Expect(err.Error()).To(Equal(`Node 'node/a/b' must cannot set 'pod' to 'children' because it is already part of another pod`))
			})
		})

		Describe("parsing valid volume configs in pods", func() {
			var (
				err    error
				appDef V2AppDefinition
			)

			BeforeEach(func() {
				byteSlice := []byte(`{
		          "nodes": {
		              "session1":
		                {
		                  "pod": "children"
		                },
		              "session1/api":
		                {
		                  "image": "registry/namespace/repository:version",
		                  "volumes": [
		                  	{"path": "/data1", "size": "27 GB"}
		                  ]
		                },
		              "session1/alt1":
		                {
		                  "image": "dockerfile/redis",
		                  "volumes": [
		                  	{"volumes-from": "session1/api"}
		                  ]
		                },
		              "session1/alt2":
		                {
		                  "image": "dockerfile/redis",
		                  "volumes": [
		                  	{"volume-from": "session1/api", "volume-path": "/data1"}
		                  ]
		                },
		              "session1/alt3":
		                {
		                  "image": "dockerfile/redis",
		                  "volumes": [
		                  	{"volume-from": "session1/api", "volume-path": "/data1", "path": "/alt4"}
		                  ]
		                }
			          }
			        }`)

				err = json.Unmarshal(byteSlice, &appDef)
			})

			It("should not throw error", func() {
				Expect(err).To(BeNil())
			})

			It("should parse 5 nodes", func() {
				Expect(appDef.Nodes).To(HaveLen(5))
			})

			It("should parse one volume for each component", func() {
				Expect(appDef.Nodes[NodeName("session1/api")].Volumes).To(HaveLen(1))
				Expect(appDef.Nodes[NodeName("session1/alt1")].Volumes).To(HaveLen(1))
				Expect(appDef.Nodes[NodeName("session1/alt2")].Volumes).To(HaveLen(1))
				Expect(appDef.Nodes[NodeName("session1/alt3")].Volumes).To(HaveLen(1))

				Expect(appDef.Nodes[NodeName("session1/api")].Volumes[0].Path).To(Equal("/data1"))
				Expect(appDef.Nodes[NodeName("session1/api")].Volumes[0].Size).To(Equal(VolumeSize("27 GB")))
				Expect(appDef.Nodes[NodeName("session1/alt1")].Volumes[0].VolumesFrom).To(Equal("session1/api"))
				Expect(appDef.Nodes[NodeName("session1/alt2")].Volumes[0].VolumeFrom).To(Equal("session1/api"))
				Expect(appDef.Nodes[NodeName("session1/alt2")].Volumes[0].VolumePath).To(Equal("/data1"))
				Expect(appDef.Nodes[NodeName("session1/alt3")].Volumes[0].VolumeFrom).To(Equal("session1/api"))
				Expect(appDef.Nodes[NodeName("session1/alt3")].Volumes[0].VolumePath).To(Equal("/data1"))
				Expect(appDef.Nodes[NodeName("session1/alt3")].Volumes[0].Path).To(Equal("/alt4"))

			})
		})

		Describe("parsing invalid volume configs in pods, invalid property combi path+volumes-from", func() {
			var err error

			BeforeEach(func() {
				nodes := testApp()
				nodes["node/a"] = setPod(testNode(), PodChildren)
				nodes["node/a/b1"] = addVols(testNode(), VolumeConfig{Path: "/data1", Size: VolumeSize("27 GB"), VolumesFrom: "node/a/b2"})
				nodes["node/a/b2"] = addVols(testNode(), VolumeConfig{VolumesFrom: "node/a/b1"})

				err = validate(nodes)
			})

			It("should throw error InvalidVolumeConfigError", func() {
				Expect(IsInvalidVolumeConfig(err)).To(BeTrue())
				Expect(err.Error()).To(Equal(`Cannot parse volume config. Volumes-from for path '/data1' should be empty.`))
			})
		})

		Describe("parsing invalid volume configs in pods, invalid property combi path+volume-from", func() {
			var err error

			BeforeEach(func() {
				nodes := testApp()
				nodes["node/a"] = setPod(testNode(), PodChildren)
				nodes["node/a/b1"] = addVols(testNode(), VolumeConfig{Path: "/data1", Size: VolumeSize("27 GB"), VolumeFrom: "node/a/b2"})
				nodes["node/a/b2"] = addVols(testNode(), VolumeConfig{VolumesFrom: "node/a/b1"})

				err = validate(nodes)
			})

			It("should throw error InvalidVolumeConfigError", func() {
				Expect(IsInvalidVolumeConfig(err)).To(BeTrue())
				Expect(err.Error()).To(Equal(`Cannot parse volume config. Volume-from for path '/data1' should be empty.`))
			})
		})

		Describe("parsing invalid volume configs in pods, invalid property combi volumes-from+path", func() {
			var err error

			BeforeEach(func() {
				nodes := testApp()
				nodes["node/a"] = setPod(testNode(), PodChildren)
				nodes["node/a/b1"] = addVols(testNode(), VolumeConfig{Path: "/data1", Size: VolumeSize("27 GB")})
				nodes["node/a/b2"] = addVols(testNode(), VolumeConfig{VolumesFrom: "node/a/b1", Path: "/alt1"})

				err = validate(nodes)
			})

			It("should throw error InvalidVolumeConfigError", func() {
				Expect(IsInvalidVolumeConfig(err)).To(BeTrue())
				Expect(err.Error()).To(Equal(`Cannot parse volume config. Path for volumes-from 'node/a/b1' should be empty.`))
			})
		})

		Describe("parsing invalid volume configs in pods, invalid property combi volumes-from+size", func() {
			var err error

			BeforeEach(func() {
				nodes := testApp()
				nodes["node/a"] = setPod(testNode(), PodChildren)
				nodes["node/a/b1"] = addVols(testNode(), VolumeConfig{Path: "/data1", Size: VolumeSize("27 GB")})
				nodes["node/a/b2"] = addVols(testNode(), VolumeConfig{VolumesFrom: "node/a/b1", Size: VolumeSize("5 GB")})

				err = validate(nodes)
			})

			It("should throw error InvalidVolumeConfigError", func() {
				Expect(IsInvalidVolumeConfig(err)).To(BeTrue())
				Expect(err.Error()).To(Equal(`Cannot parse volume config. Size for volumes-from 'node/a/b1' should be empty.`))
			})
		})

		Describe("parsing invalid volume configs in pods, invalid property combi volumes-from+volume-from", func() {
			var err error

			BeforeEach(func() {
				nodes := testApp()
				nodes["node/a"] = setPod(testNode(), PodChildren)
				nodes["node/a/b1"] = addVols(testNode(), VolumeConfig{Path: "/data1", Size: VolumeSize("27 GB")})
				nodes["node/a/b2"] = addVols(testNode(), VolumeConfig{VolumesFrom: "node/a/b1", VolumeFrom: "node/a/b1"})

				err = validate(nodes)
			})

			It("should throw error InvalidVolumeConfigError", func() {
				Expect(IsInvalidVolumeConfig(err)).To(BeTrue())
				Expect(err.Error()).To(Equal(`Cannot parse volume config. Volume-from for volumes-from 'node/a/b1' should be empty.`))
			})
		})

		Describe("parsing invalid volume configs in pods, invalid property combi volume-from+size", func() {
			var err error

			BeforeEach(func() {
				nodes := testApp()
				nodes["node/a"] = setPod(testNode(), PodChildren)
				nodes["node/a/b1"] = addVols(testNode(), VolumeConfig{Path: "/data1", Size: VolumeSize("27 GB")})
				nodes["node/a/b2"] = addVols(testNode(), VolumeConfig{VolumeFrom: "node/a/b1", VolumePath: "/data1", Size: VolumeSize("5GB")})

				err = validate(nodes)
			})

			It("should throw error InvalidVolumeConfigError", func() {
				Expect(IsInvalidVolumeConfig(err)).To(BeTrue())
				Expect(err.Error()).To(Equal(`Cannot parse volume config. Size for volume-from 'node/a/b1' should be empty.`))
			})
		})

		Describe("parsing invalid volume configs in pods, volumes-from cannot refer to self", func() {
			var err error

			BeforeEach(func() {
				nodes := testApp()
				nodes["node/a"] = setPod(testNode(), PodChildren)
				nodes["node/a/b1"] = addVols(testNode(), VolumeConfig{Path: "/data1", Size: VolumeSize("27 GB")})
				nodes["node/a/b2"] = addVols(testNode(), VolumeConfig{VolumesFrom: "node/a/b2"})

				err = validate(nodes)
			})

			It("should throw error InvalidVolumeConfigError", func() {
				Expect(IsInvalidVolumeConfig(err)).To(BeTrue())
				Expect(err.Error()).To(Equal(`Cannot parse volume config. Cannot refer to own node 'node/a/b2'.`))
			})
		})

		Describe("parsing invalid volume configs in pod, volume-from cannot refer to self", func() {
			var err error

			BeforeEach(func() {
				nodes := testApp()
				nodes["node/a"] = setPod(testNode(), PodChildren)
				nodes["node/a/b1"] = addVols(testNode(), VolumeConfig{Path: "/data1", Size: VolumeSize("27 GB")})
				nodes["node/a/b2"] = addVols(testNode(), VolumeConfig{VolumeFrom: "node/a/b2", VolumePath: "/data1"})

				err = validate(nodes)
			})

			It("should throw error InvalidVolumeConfigError", func() {
				Expect(IsInvalidVolumeConfig(err)).To(BeTrue())
				Expect(err.Error()).To(Equal(`Cannot parse volume config. Cannot refer to own node 'node/a/b2'.`))
			})
		})

		Describe("parsing invalid volume configs in pods, unknown path in volume-path", func() {
			var err error

			BeforeEach(func() {
				nodes := testApp()
				nodes["node/a"] = setPod(testNode(), PodChildren)
				nodes["node/a/b1"] = addVols(testNode(), VolumeConfig{Path: "/data1", Size: VolumeSize("27 GB")})
				nodes["node/a/b2"] = addVols(testNode(), VolumeConfig{VolumeFrom: "node/a/b1", VolumePath: "/unknown"})

				err = validate(nodes)
			})

			It("should throw error InvalidVolumeConfigError", func() {
				Expect(IsInvalidVolumeConfig(err)).To(BeTrue())
				Expect(err.Error()).To(Equal(`Cannot parse volume config. Cannot find path '/unknown' on node 'node/a/b1'.`))
			})
		})

		Describe("parsing invalid volume configs, duplicate volume via different postfixes", func() {
			var err error

			BeforeEach(func() {
				nodes := testApp()
				nodes["node/a"] = addVols(testNode(),
					VolumeConfig{Path: "/xdata1", Size: VolumeSize("27 GB")},
					VolumeConfig{Path: "/xdata1/", Size: VolumeSize("27 GB")},
				)

				err = validate(nodes)
			})

			It("should throw error DuplicateVolumePathError", func() {
				Expect(IsDuplicateVolumePath(err)).To(BeTrue())
				Expect(err.Error()).To(Equal(`Cannot parse app config. Duplicate volume '/xdata1' found in node 'node/a'.`))
			})
		})

		Describe("parsing invalid volume configs in pods, duplicate volume via volumes-from", func() {
			var err error

			BeforeEach(func() {
				nodes := testApp()
				nodes["node/a"] = setPod(testNode(), PodChildren)
				nodes["node/a/b1"] = addVols(testNode(), VolumeConfig{Path: "/xdata1", Size: VolumeSize("27 GB")})
				nodes["node/a/b2"] = addVols(testNode(),
					VolumeConfig{VolumesFrom: "node/a/b1"},
					VolumeConfig{Path: "/xdata1", Size: VolumeSize("5GB")},
				)

				err = validate(nodes)
			})

			It("should throw error DuplicateVolumePathError", func() {
				//Expect(IsDuplicateVolumePath(err)).To(BeTrue())
				Expect(err.Error()).To(Equal(`Cannot parse app config. Duplicate volume '/xdata1' found in node 'node/a/b2'.`))
			})
		})

		Describe("parsing invalid volume configs in pods, duplicate volume via volume-from", func() {
			var err error

			BeforeEach(func() {
				nodes := testApp()
				nodes["node/a"] = setPod(testNode(), PodChildren)
				nodes["node/a/b1"] = addVols(testNode(), VolumeConfig{Path: "/xdata1", Size: VolumeSize("27 GB")})
				nodes["node/a/b2"] = addVols(testNode(),
					VolumeConfig{Path: "/xdata1", Size: VolumeSize("5GB")},
					VolumeConfig{VolumeFrom: "node/a/b1", VolumePath: "/xdata1"},
				)

				err = validate(nodes)
			})

			It("should throw error DuplicateVolumePathError", func() {
				Expect(IsDuplicateVolumePath(err)).To(BeTrue())
				Expect(err.Error()).To(Equal(`Cannot parse app config. Duplicate volume '/xdata1' found in node 'node/a/b2'.`))
			})
		})

		Describe("parsing invalid volume configs in pods, duplicate volume via linked volumes-from", func() {
			var err error

			BeforeEach(func() {
				nodes := testApp()
				nodes["node/a"] = setPod(testNode(), PodChildren)
				nodes["node/a/b1"] = addVols(testNode(), VolumeConfig{Path: "/xdata1", Size: VolumeSize("27 GB")})
				nodes["node/a/b2"] = addVols(testNode(), VolumeConfig{VolumesFrom: "node/a/b1"})
				nodes["node/a/b3"] = addVols(testNode(),
					VolumeConfig{VolumesFrom: "node/a/b1"},
					VolumeConfig{VolumesFrom: "node/a/b2"},
				)

				err = validate(nodes)
			})

			It("should throw error DuplicateVolumePathError", func() {
				Expect(IsDuplicateVolumePath(err)).To(BeTrue())
				Expect(err.Error()).To(Equal(`Cannot parse app config. Duplicate volume '/xdata1' found in node 'node/a/b3'.`))
			})
		})

		Describe("parsing invalid volume configs in pods, cycle in volumes-from references", func() {
			var err error

			BeforeEach(func() {
				nodes := testApp()
				nodes["node/a"] = setPod(testNode(), PodChildren)
				nodes["node/a/b1"] = addVols(testNode(), VolumeConfig{VolumesFrom: "node/a/b3"})
				nodes["node/a/b2"] = addVols(testNode(), VolumeConfig{VolumesFrom: "node/a/b1"})
				nodes["node/a/b3"] = addVols(testNode(), VolumeConfig{VolumesFrom: "node/a/b2"})

				err = validate(nodes)
			})

			It("should throw error VolumeCycleError", func() {
				Expect(IsVolumeCycle(err)).To(BeTrue())
			})
		})

		Describe("parsing dependency configs in pods, same name should result in same port", func() {
			var err error

			BeforeEach(func() {
				nodes := testApp()
				nodes["node/a"] = setPod(testNode(), PodChildren)
				nodes["node/a/b1"] = addDeps(testNode(), DependencyConfig{Name: "redis", Port: generictypes.MustParseDockerPort("6379")})
				nodes["node/a/b2"] = addDeps(testNode(), DependencyConfig{Name: "redis", Port: generictypes.MustParseDockerPort("6379")})
				nodes["redis"] = addPorts(testNode(), generictypes.MustParseDockerPort("6379"))

				err = validate(nodes)
			})

			It("should not throw an error", func() {
				Expect(err).To(BeNil())
			})
		})

		Describe("parsing invalid dependency configs in pods, same name different ports", func() {
			var err error

			BeforeEach(func() {
				nodes := testApp()
				nodes["node/a"] = setPod(testNode(), PodChildren)
				nodes["node/a/b1"] = addDeps(testNode(), DependencyConfig{Name: "redis1", Port: generictypes.MustParseDockerPort("6379")})
				nodes["node/a/b2"] = addDeps(testNode(), DependencyConfig{Name: "redis1", Port: generictypes.MustParseDockerPort("1234")})
				nodes["redis1"] = addPorts(testNode(), generictypes.MustParseDockerPort("6379"), generictypes.MustParseDockerPort("1234"))
				nodes["redis2"] = addPorts(testNode(), generictypes.MustParseDockerPort("6379"))

				err = validate(nodes)
			})

			It("should throw error InvalidDependencyConfigError", func() {
				Expect(IsInvalidDependencyConfig(err)).To(BeTrue())
				Expect(err.Error()).To(Equal(`Cannot parse app config. Duplicate (but different ports) dependency 'redis1' in pod under 'node/a'.`))
			})
		})

		Describe("parsing invalid dependency configs in pods, same alias different names", func() {
			var err error

			BeforeEach(func() {
				nodes := testApp()
				nodes["node/a"] = setPod(testNode(), PodChildren)
				nodes["node/a/b1"] = addDeps(testNode(), DependencyConfig{Alias: "db", Name: "redis1", Port: generictypes.MustParseDockerPort("6379")})
				nodes["node/a/b2"] = addDeps(testNode(), DependencyConfig{Alias: "db", Name: "redis2", Port: generictypes.MustParseDockerPort("6379")})
				nodes["redis1"] = addPorts(testNode(), generictypes.MustParseDockerPort("6379"))
				nodes["redis2"] = addPorts(testNode(), generictypes.MustParseDockerPort("6379"))

				err = validate(nodes)
			})

			It("should throw error InvalidDependencyConfigError", func() {
				Expect(IsInvalidDependencyConfig(err)).To(BeTrue())
				Expect(err.Error()).To(Equal(`Cannot parse app config. Duplicate (but different names) dependency 'db' in pod under 'node/a'.`))
			})
		})

		Describe("parsing invalid dependency configs in pods, one with aliases, other with (conflicting) name", func() {
			var err error

			BeforeEach(func() {
				nodes := testApp()
				nodes["node/a"] = setPod(testNode(), PodChildren)
				nodes["node/a/b1"] = addDeps(testNode(), DependencyConfig{Name: "storage/redis", Port: generictypes.MustParseDockerPort("6379")})
				nodes["node/a/b2"] = addDeps(testNode(), DependencyConfig{Alias: "redis", Name: "storage/redis2", Port: generictypes.MustParseDockerPort("6379")})
				nodes["redis1"] = addPorts(testNode(), generictypes.MustParseDockerPort("6379"))
				nodes["redis2"] = addPorts(testNode(), generictypes.MustParseDockerPort("6379"))
				nodes["storage/redis"] = addPorts(testNode(), generictypes.MustParseDockerPort("6379"))
				nodes["storage/redis2"] = addPorts(testNode(), generictypes.MustParseDockerPort("6379"))

				err = validate(nodes)
			})

			It("should throw error InvalidDependencyConfigError", func() {
				Expect(IsInvalidDependencyConfig(err)).To(BeTrue())
				Expect(err.Error()).To(Equal(`Cannot parse app config. Duplicate (but different names) dependency 'redis' in pod under 'node/a'.`))
			})
		})

		Describe("parsing invalid dependency configs in pods, same alias different ports", func() {
			var err error

			BeforeEach(func() {
				nodes := testApp()
				nodes["node/a"] = setPod(testNode(), PodChildren)
				nodes["node/a/b1"] = addDeps(testNode(), DependencyConfig{Alias: "db", Name: "redis", Port: generictypes.MustParseDockerPort("6379")})
				nodes["node/a/b2"] = addDeps(testNode(), DependencyConfig{Alias: "db", Name: "redis", Port: generictypes.MustParseDockerPort("9736")})
				nodes["redis"] = addPorts(testNode(), generictypes.MustParseDockerPort("6379"), generictypes.MustParseDockerPort("9736"))
				nodes["redis2"] = addPorts(testNode(), generictypes.MustParseDockerPort("6379"))

				err = validate(nodes)
			})

			It("should throw error InvalidDependencyConfigError", func() {
				Expect(IsInvalidDependencyConfig(err)).To(BeTrue())
				Expect(err.Error()).To(Equal(`Cannot parse app config. Duplicate (but different ports) dependency 'db' in pod under 'node/a'.`))
			})
		})

		Describe("parsing valid scaling configs in pods, scaling values should be the same in all nodes of a pod or not set", func() {
			var err error

			BeforeEach(func() {
				nodes := testApp()
				nodes["node/a"] = setPod(testNode(), PodChildren)
				nodes["node/a/b1"] = addScale(testNode(), 1, maxScale)
				nodes["node/a/b2"] = addScale(testNode(), 0, maxScale)
				nodes["node/a/b3"] = addScale(testNode(), 1, 0)
				nodes["node/a/b4"] = testNode()

				err = validate(nodes)
			})

			It("should not throw an error", func() {
				Expect(err).To(BeNil())
			})
		})

		Describe("parsing invalid scaling definition in pods, minimum scaling values should be the same in all nodes of a pod", func() {
			var err error

			BeforeEach(func() {
				nodes := testApp()
				nodes["node/a"] = setPod(testNode(), PodChildren)
				nodes["node/a/b1"] = addScale(testNode(), 1, 5)
				nodes["node/a/b2"] = addScale(testNode(), 2, 5)

				err = validate(nodes)
			})

			It("should throw error InvalidScalingConfigError", func() {
				Expect(IsInvalidScalingConfig(err)).To(BeTrue())
				Expect(err.Error()).To(Equal(`Cannot parse app config. Different minimum scaling policies in pod under 'node/a'.`))
			})
		})

		Describe("parsing invalid scaling definition in pods, maximum scaling values should be the same in all components of a pod", func() {
			var err error

			BeforeEach(func() {
				nodes := testApp()
				nodes["node/a"] = setPod(testNode(), PodChildren)
				nodes["node/a/b1"] = addScale(testNode(), 2, 5)
				nodes["node/a/b2"] = addScale(testNode(), 2, 7)

				err = validate(nodes)
			})

			It("should throw error InvalidScalingConfigError", func() {
				Expect(IsInvalidScalingConfig(err)).To(BeTrue())
				Expect(err.Error()).To(Equal(`Cannot parse app config. Different maximum scaling policies in pod under 'node/a'.`))
			})
		})

		Describe("parsing invalid ports configs in pods, cannot duplicate ports in a single pod", func() {
			var err error

			BeforeEach(func() {
				nodes := testApp()
				nodes["node/a"] = setPod(testNode(), PodChildren)
				nodes["node/a/b1"] = addPorts(testNode(), generictypes.MustParseDockerPort("80"))
				nodes["node/a/b2"] = addPorts(testNode(), generictypes.MustParseDockerPort("80"))

				err = validate(nodes)
			})

			It("should throw error InvalidPortConfigError", func() {
				Expect(IsInvalidPortConfig(err)).To(BeTrue())
				Expect(err.Error()).To(Equal(`Cannot parse app config. Multiple nodes export port '80/tcp' in pod under 'node/a'.`))
			})
		})

		Describe("parsing invalid ports configs in pods, cannot duplicate ports in a single pod (mixed with/without protocol)", func() {
			var err error

			BeforeEach(func() {
				nodes := testApp()
				nodes["node/a"] = setPod(testNode(), PodChildren)
				nodes["node/a/b1"] = addPorts(testNode(), generictypes.MustParseDockerPort("80"))
				nodes["node/a/b2"] = addPorts(testNode(), generictypes.MustParseDockerPort("80/tcp"))

				err = validate(nodes)
			})

			It("should throw error InvalidPortConfigError", func() {
				Expect(IsInvalidPortConfig(err)).To(BeTrue())
				Expect(err.Error()).To(Equal(`Cannot parse app config. Multiple nodes export port '80/tcp' in pod under 'node/a'.`))
			})
		})

	})
})
