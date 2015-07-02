package userconfig

import (
	"testing"

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

	/*		addLinks := func(config *NodeDefinition, dep ...DependencyConfig) *NodeDefinition {
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
			}*/

	testApp := func() NodeDefinitions {
		return make(NodeDefinitions)
	}

	valCtx := func() *ValidationContext {
		return &ValidationContext{
			Protocols:     []string{"tcp"},
			MinVolumeSize: "1 GB",
			MaxVolumeSize: "100 GB",
		}
	}

	Describe("pod tests ", func() {
		Describe("parsing (valid) pod==children on a node that has 2 children", func() {
			var err error

			BeforeEach(func() {
				nodes := testApp()
				nodes["node/a"] = setPod(testNode(), PodChildren)
				nodes["node/a/b"] = testNode()
				nodes["node/a/c"] = testNode()

				err = nodes.validate(valCtx())
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

				err = nodes.validate(valCtx())
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

				err = nodes.validate(valCtx())
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

				err = nodes.validate(valCtx())
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

				err = nodes.validate(valCtx())
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

				err = nodes.validate(valCtx())
			})

			It("should throw error IsInvalidPodConfig", func() {
				Expect(IsInvalidPodConfig(err)).To(BeTrue())
				Expect(err.Error()).To(Equal(`Node 'node/a/b' must cannot set 'pod' to 'children' because it is already part of another pod`))
			})

		})

		/*
			Describe("parsing valid volume configs in pods", func() {
				var (
					err       error
					byteSlice []byte
					appConfig AppDefinition
				)

				BeforeEach(func() {
					byteSlice = []byte(`{
			          "app_name": "test-app-name",
			          "services": [
			            {
			              "service_name": "session1",
			              "components": [
			                {
			                  "component_name": "api",
			                  "image": "registry/namespace/repository:version",
			                  "pod": "ns4",
			                  "volumes": [
			                  	{"path": "/data1", "size": "27 GB"}
			                  ]
			                },
			                {
			                  "component_name": "alt1",
			                  "image": "dockerfile/redis",
			                  "pod": "ns4",
			                  "volumes": [
			                  	{"volumes-from": "api"}
			                  ]
			                },
			                {
			                  "component_name": "alt2",
			                  "image": "dockerfile/redis",
			                  "pod": "ns4",
			                  "volumes": [
			                  	{"volume-from": "api", "volume-path": "/data1"}
			                  ]
			                },
			                {
			                  "component_name": "alt3",
			                  "image": "dockerfile/redis",
			                  "pod": "ns4",
			                  "volumes": [
			                  	{"volume-from": "api", "volume-path": "/data1", "path": "/alt4"}
			                  ]
			                }
			              ]
			            }
			          ]
			        }`)

					err = json.Unmarshal(byteSlice, &appConfig)
				})

				It("should not throw error", func() {
					Expect(err).To(BeNil())
				})

				It("should parse one service", func() {
					Expect(appConfig.Services).To(HaveLen(1))
				})

				It("should parse 4 components", func() {
					Expect(appConfig.Services[0].Components).To(HaveLen(4))
				})

				It("should parse one volume for each component", func() {
					Expect(appConfig.Services[0].Components[0].Volumes).To(HaveLen(1))
					Expect(appConfig.Services[0].Components[1].Volumes).To(HaveLen(1))
					Expect(appConfig.Services[0].Components[2].Volumes).To(HaveLen(1))
					Expect(appConfig.Services[0].Components[3].Volumes).To(HaveLen(1))

					Expect(appConfig.Services[0].Components[0].Volumes[0].Path).To(Equal("/data1"))
					Expect(appConfig.Services[0].Components[0].Volumes[0].Size).To(Equal(VolumeSize("27 GB")))
					Expect(appConfig.Services[0].Components[1].Volumes[0].VolumesFrom).To(Equal("api"))
					Expect(appConfig.Services[0].Components[2].Volumes[0].VolumeFrom).To(Equal("api"))
					Expect(appConfig.Services[0].Components[2].Volumes[0].VolumePath).To(Equal("/data1"))
					Expect(appConfig.Services[0].Components[3].Volumes[0].VolumeFrom).To(Equal("api"))
					Expect(appConfig.Services[0].Components[3].Volumes[0].VolumePath).To(Equal("/data1"))
					Expect(appConfig.Services[0].Components[3].Volumes[0].Path).To(Equal("/alt4"))

				})

			})
		*/

		Describe("parsing invalid volume configs in pods, invalid property combi path+volumes-from", func() {
			var err error

			BeforeEach(func() {
				nodes := testApp()
				nodes["node/a"] = setPod(testNode(), PodChildren)
				nodes["node/a/b1"] = addVols(testNode(), VolumeConfig{Path: "/data1", Size: VolumeSize("27 GB"), VolumesFrom: "node/a/b2"})
				nodes["node/a/b2"] = addVols(testNode(), VolumeConfig{VolumesFrom: "node/a/b1"})

				err = nodes.validate(valCtx())
			})

			It("should throw error InvalidVolumeConfigError", func() {
				Expect(IsInvalidVolumeConfig(err)).To(BeTrue())
				Expect(err.Error()).To(Equal(`Cannot parse volume config. Volumes-from for path '/data1' should be empty.`))
			})

		})

		/*
			Describe("parsing invalid volume configs in pods, invalid property combi path+volume-from", func() {
				var err error

				BeforeEach(func() {
					appConfig := testApp(
						testService("session1",
							addVols(testComponent("api", "ns4"), VolumeConfig{Path: "/data1", Size: VolumeSize("27 GB"), VolumeFrom: "alt1"}),
							addVols(testComponent("alt1", "ns4"), VolumeConfig{VolumesFrom: "api"}),
						),
					)

					err = appConfig.validate()
				})

				It("should throw error InvalidVolumeConfigError", func() {
					Expect(IsInvalidVolumeConfig(err)).To(BeTrue())
					Expect(err.Error()).To(Equal(`Cannot parse volume config. Volume-from for path '/data1' should be empty.`))
				})

			})

			Describe("parsing invalid volume configs in pods, invalid property combi volumes-from+path", func() {
				var err error

				BeforeEach(func() {
					appConfig := testApp(
						testService("session1",
							addVols(testComponent("api", "ns4"), VolumeConfig{Path: "/data1", Size: VolumeSize("27 GB")}),
							addVols(testComponent("alt1", "ns4"), VolumeConfig{VolumesFrom: "api", Path: "/alt1"}),
						),
					)

					err = appConfig.validate()
				})

				It("should throw error InvalidVolumeConfigError", func() {
					Expect(IsInvalidVolumeConfig(err)).To(BeTrue())
					Expect(err.Error()).To(Equal(`Cannot parse volume config. Path for volumes-from 'api' should be empty.`))
				})

			})

			Describe("parsing invalid volume configs in pods, invalid property combi volumes-from+size", func() {
				var err error

				BeforeEach(func() {
					appConfig := testApp(
						testService("session1",
							addVols(testComponent("api", "ns4"), VolumeConfig{Path: "/data1", Size: VolumeSize("27 GB")}),
							addVols(testComponent("alt1", "ns4"), VolumeConfig{VolumesFrom: "api", Size: VolumeSize("5 GB")}),
						),
					)

					err = appConfig.validate()
				})

				It("should throw error InvalidVolumeConfigError", func() {
					Expect(IsInvalidVolumeConfig(err)).To(BeTrue())
					Expect(err.Error()).To(Equal(`Cannot parse volume config. Size for volumes-from 'api' should be empty.`))
				})

			})

			Describe("parsing invalid volume configs in pods, invalid property combi volumes-from+volume-from", func() {
				var err error

				BeforeEach(func() {
					appConfig := testApp(
						testService("session1",
							addVols(testComponent("api", "ns4"), VolumeConfig{Path: "/data1", Size: VolumeSize("27 GB")}),
							addVols(testComponent("alt1", "ns4"), VolumeConfig{VolumesFrom: "api", VolumeFrom: "api"}),
						),
					)

					err = appConfig.validate()
				})

				It("should throw error InvalidVolumeConfigError", func() {
					Expect(IsInvalidVolumeConfig(err)).To(BeTrue())
					Expect(err.Error()).To(Equal(`Cannot parse volume config. Volume-from for volumes-from 'api' should be empty.`))
				})

			})

			Describe("parsing invalid volume configs in pods, invalid property combi volume-from+size", func() {
				var err error

				BeforeEach(func() {
					appConfig := testApp(
						testService("session1",
							addVols(testComponent("api", "ns4"), VolumeConfig{Path: "/data1", Size: VolumeSize("27 GB")}),
							addVols(testComponent("alt1", "ns4"), VolumeConfig{VolumeFrom: "api", VolumePath: "/data1", Size: VolumeSize("5GB")}),
						),
					)

					err = appConfig.validate()
				})

				It("should throw error InvalidVolumeConfigError", func() {
					Expect(IsInvalidVolumeConfig(err)).To(BeTrue())
					Expect(err.Error()).To(Equal(`Cannot parse volume config. Size for volume-from 'api' should be empty.`))
				})

			})

			Describe("parsing invalid volume configs in pods, volumes-from cannot refer to self", func() {
				var err error

				BeforeEach(func() {
					appConfig := testApp(
						testService("session1",
							addVols(testComponent("api", "ns4"), VolumeConfig{Path: "/data1", Size: VolumeSize("27 GB")}),
							addVols(testComponent("alt1", "ns4"), VolumeConfig{VolumesFrom: "alt1"}),
						),
					)

					err = appConfig.validate()
				})

				It("should throw error InvalidVolumeConfigError", func() {
					Expect(IsInvalidVolumeConfig(err)).To(BeTrue())
					Expect(err.Error()).To(Equal(`Cannot parse volume config. Cannot refer to own component 'alt1'.`))
				})

			})

			Describe("parsing invalid volume configs in pod, volume-from cannot refer to self", func() {
				var err error

				BeforeEach(func() {
					appConfig := testApp(
						testService("session1",
							addVols(testComponent("api", "ns4"), VolumeConfig{Path: "/data1", Size: VolumeSize("27 GB")}),
							addVols(testComponent("alt1", "ns4"), VolumeConfig{VolumeFrom: "alt1", VolumePath: "/data1"}),
						),
					)

					err = appConfig.validate()
				})

				It("should throw error InvalidVolumeConfigError", func() {
					Expect(IsInvalidVolumeConfig(err)).To(BeTrue())
					Expect(err.Error()).To(Equal(`Cannot parse volume config. Cannot refer to own component 'alt1'.`))
				})

			})

			Describe("parsing invalid volume configs in pods, unknown path in volume-path", func() {
				var err error

				BeforeEach(func() {
					appConfig := testApp(
						testService("session1",
							addVols(testComponent("api", "ns4"), VolumeConfig{Path: "/data1", Size: VolumeSize("27 GB")}),
							addVols(testComponent("alt1", "ns4"), VolumeConfig{VolumeFrom: "api", VolumePath: "/unknown"}),
						),
					)

					err = appConfig.validate()
				})

				It("should throw error InvalidVolumeConfigError", func() {
					Expect(IsInvalidVolumeConfig(err)).To(BeTrue())
					Expect(err.Error()).To(Equal(`Cannot parse volume config. Cannot find path '/unknown' on component 'api'.`))
				})

			})

			Describe("parsing invalid volume configs, duplicate volume via different postfixes", func() {
				var err error

				BeforeEach(func() {
					appConfig := testApp(
						testService("session1",
							addVols(testComponent("api", ""),
								VolumeConfig{Path: "/xdata1", Size: VolumeSize("27 GB")},
								VolumeConfig{Path: "/xdata1/", Size: VolumeSize("27 GB")},
							),
						),
					)

					err = appConfig.validate()
				})

				It("should throw error DuplicateVolumePathError", func() {
					Expect(IsDuplicateVolumePath(err)).To(BeTrue())
					Expect(err.Error()).To(Equal(`Cannot parse app config. Duplicate volume '/xdata1' found in component 'api'.`))
				})

			})

			Describe("parsing invalid volume configs in pods, duplicate volume via volumes-from", func() {
				var err error

				BeforeEach(func() {
					appConfig := testApp(
						testService("session1",
							addVols(testComponent("api", "ns4"), VolumeConfig{Path: "/xdata1", Size: VolumeSize("27 GB")}),
							addVols(testComponent("alt1", "ns4"),
								VolumeConfig{VolumesFrom: "api"},
								VolumeConfig{Path: "/xdata1", Size: VolumeSize("5GB")},
							),
						),
					)

					err = appConfig.validate()
				})

				It("should throw error DuplicateVolumePathError", func() {
					Expect(IsDuplicateVolumePath(err)).To(BeTrue())
					Expect(err.Error()).To(Equal(`Cannot parse app config. Duplicate volume '/xdata1' found in component 'alt1'.`))
				})

			})

			Describe("parsing invalid volume configs in pods, duplicate volume via volume-from", func() {
				var err error

				BeforeEach(func() {
					appConfig := testApp(
						testService("session1",
							addVols(testComponent("api", "ns4"), VolumeConfig{Path: "/xdata1", Size: VolumeSize("27 GB")}),
							addVols(testComponent("alt1", "ns4"),
								VolumeConfig{Path: "/xdata1", Size: VolumeSize("5GB")},
								VolumeConfig{VolumeFrom: "api", VolumePath: "/xdata1"},
							),
						),
					)

					err = appConfig.validate()
				})

				It("should throw error DuplicateVolumePathError", func() {
					Expect(IsDuplicateVolumePath(err)).To(BeTrue())
					Expect(err.Error()).To(Equal(`Cannot parse app config. Duplicate volume '/xdata1' found in component 'alt1'.`))
				})

			})

			Describe("parsing invalid volume configs in pods, duplicate volume via linked volumes-from", func() {
				var err error

				BeforeEach(func() {
					appConfig := testApp(
						testService("session1",
							addVols(testComponent("api", "ns4"), VolumeConfig{Path: "/xdata1", Size: VolumeSize("27 GB")}),
							addVols(testComponent("alt1", "ns4"), VolumeConfig{VolumesFrom: "api"}),
							addVols(testComponent("alt2", "ns4"),
								VolumeConfig{VolumesFrom: "api"},
								VolumeConfig{VolumesFrom: "alt1"},
							),
						),
					)

					err = appConfig.validate()
				})

				It("should throw error DuplicateVolumePathError", func() {
					Expect(IsDuplicateVolumePath(err)).To(BeTrue())
					Expect(err.Error()).To(Equal(`Cannot parse app config. Duplicate volume '/xdata1' found in component 'alt2'.`))
				})

			})

			Describe("parsing invalid volume configs in pods, cycle in volumes-from references", func() {
				var err error

				BeforeEach(func() {
					appConfig := testApp(
						testService("session1",
							addVols(testComponent("api", "ns4"), VolumeConfig{VolumesFrom: "alt2"}),
							addVols(testComponent("alt1", "ns4"), VolumeConfig{VolumesFrom: "api"}),
							addVols(testComponent("alt2", "ns4"), VolumeConfig{VolumesFrom: "alt1"}),
						),
					)

					err = appConfig.validate()
				})

				It("should throw error InvalidVolumeConfigError", func() {
					Expect(IsInvalidVolumeConfig(err)).To(BeTrue())
					Expect(err.Error()).To(Equal(`Cannot parse app config. Cycle in referenced components detected in 'alt2'.`))
				})

			})

			Describe("parsing dependency configs in pods, same name should result in same port", func() {
				var err error

				BeforeEach(func() {
					appConfig := testApp(
						testService("session1",
							addDeps(testComponent("alt1", "ns4"), DependencyConfig{Name: "redis", Port: generictypes.MustParseDockerPort("6379")}),
							addDeps(testComponent("alt2", "ns4"), DependencyConfig{Name: "redis", Port: generictypes.MustParseDockerPort("6379")}),
							testComponent("redis", ""),
						),
					)

					err = appConfig.validate()
				})

				It("should not throw an error", func() {
					Expect(err).To(BeNil())
				})
			})

			Describe("parsing invalid dependency configs in pods, same name different ports", func() {
				var err error

				BeforeEach(func() {
					appConfig := testApp(
						testService("session1",
							addDeps(testComponent("alt1", "ns4"), DependencyConfig{Name: "redis1", Port: generictypes.MustParseDockerPort("6379")}),
							addDeps(testComponent("alt2", "ns4"), DependencyConfig{Name: "redis1", Port: generictypes.MustParseDockerPort("1234")}),
							testComponent("redis1", ""),
							testComponent("redis2", ""),
						),
					)

					err = appConfig.validate()
				})

				It("should throw error InvalidDependencyConfigError", func() {
					Expect(IsInvalidDependencyConfig(err)).To(BeTrue())
					Expect(err.Error()).To(Equal(`Cannot parse app config. Duplicate (but different ports) dependency 'redis1' in pod 'ns4'.`))
				})
			})

			Describe("parsing invalid dependency configs in pods, same alias different names", func() {
				var err error

				BeforeEach(func() {
					appConfig := testApp(
						testService("session1",
							addDeps(testComponent("alt1", "ns4"), DependencyConfig{Alias: "db", Name: "redis1", Port: generictypes.MustParseDockerPort("6379")}),
							addDeps(testComponent("alt2", "ns4"), DependencyConfig{Alias: "db", Name: "redis2", Port: generictypes.MustParseDockerPort("6379")}),
							testComponent("redis1", ""),
							testComponent("redis2", ""),
						),
					)

					err = appConfig.validate()
				})

				It("should throw error InvalidDependencyConfigError", func() {
					Expect(IsInvalidDependencyConfig(err)).To(BeTrue())
					Expect(err.Error()).To(Equal(`Cannot parse app config. Duplicate (but different names) dependency 'db' in pod 'ns4'.`))
				})
			})

			Describe("parsing invalid dependency configs in pods, one with aliases, other with (conflicting) name", func() {
				var err error

				BeforeEach(func() {
					appConfig := testApp(
						testService("session1",
							addDeps(testComponent("alt1", "ns4"), DependencyConfig{Name: "storage/redis", Port: generictypes.MustParseDockerPort("6379")}),
							addDeps(testComponent("alt2", "ns4"), DependencyConfig{Alias: "redis", Name: "storage/redis2", Port: generictypes.MustParseDockerPort("6379")}),
							testComponent("redis1", ""),
							testComponent("redis2", ""),
						),
						testService("storage",
							testComponent("redis", ""),
							testComponent("redis2", ""),
						),
					)

					err = appConfig.validate()
				})

				It("should throw error InvalidDependencyConfigError", func() {
					Expect(IsInvalidDependencyConfig(err)).To(BeTrue())
					Expect(err.Error()).To(Equal(`Cannot parse app config. Duplicate (but different names) dependency 'redis' in pod 'ns4'.`))
				})
			})

			Describe("parsing invalid dependency configs in pods, same alias different ports", func() {
				var err error

				BeforeEach(func() {
					appConfig := testApp(
						testService("session1",
							addDeps(testComponent("alt1", "ns4"), DependencyConfig{Alias: "db", Name: "redis", Port: generictypes.MustParseDockerPort("6379")}),
							addDeps(testComponent("alt2", "ns4"), DependencyConfig{Alias: "db", Name: "redis", Port: generictypes.MustParseDockerPort("9736")}),
							testComponent("redis1", ""),
							testComponent("redis2", ""),
						),
					)

					err = appConfig.validate()
				})

				It("should throw error InvalidDependencyConfigError", func() {
					Expect(IsInvalidDependencyConfig(err)).To(BeTrue())
					Expect(err.Error()).To(Equal(`Cannot parse app config. Duplicate (but different ports) dependency 'db' in pod 'ns4'.`))
				})
			})

			Describe("parsing valid scaling configs in pods, scaling values should be the same in all components of a pod or not set", func() {
				var err error

				BeforeEach(func() {
					appConfig := testApp(
						testService("session1",
							addScale(testComponent("alt1", "ns4"), 1, 5),
							addScale(testComponent("alt2", "ns4"), 0, 5),
							addScale(testComponent("alt3", "ns4"), 1, 0),
							testComponent("alt4", "ns4"),
						),
					)

					err = appConfig.validate()
				})

				It("should not throw an error", func() {
					Expect(err).To(BeNil())
				})

			})

			Describe("parsing invalid scaling configs in pods, minimum scaling values should be the same in all components of a pod", func() {
				var err error

				BeforeEach(func() {
					appConfig := testApp(
						testService("session1",
							addScale(testComponent("alt1", "ns4"), 1, 5),
							addScale(testComponent("alt2", "ns4"), 2, 5),
						),
					)

					err = appConfig.validate()
				})

				It("should throw error InvalidScalingConfigError", func() {
					Expect(IsInvalidScalingConfig(err)).To(BeTrue())
					Expect(err.Error()).To(Equal(`Cannot parse app config. Different minimum scaling policies in pod 'ns4'.`))
				})

			})

			Describe("parsing invalid scaling configs in pods, maximum scaling values should be the same in all components of a pod", func() {
				var err error

				BeforeEach(func() {
					appConfig := testApp(
						testService("session1",
							addScale(testComponent("alt1", "ns4"), 2, 5),
							addScale(testComponent("alt2", "ns4"), 2, 7),
						),
					)

					err = appConfig.validate()
				})

				It("should throw error InvalidScalingConfigError", func() {
					Expect(IsInvalidScalingConfig(err)).To(BeTrue())
					Expect(err.Error()).To(Equal(`Cannot parse app config. Different maximum scaling policies in pod 'ns4'.`))
				})

			})

			Describe("parsing invalid ports configs in pods, cannot duplicate ports in a single pod", func() {
				var err error

				BeforeEach(func() {
					appConfig := testApp(
						testService("session1",
							addPorts(testComponent("alt1", "ns5"), generictypes.MustParseDockerPort("80")),
							addPorts(testComponent("alt2", "ns5"), generictypes.MustParseDockerPort("80")),
						),
					)

					err = appConfig.validate()
				})

				It("should throw error InvalidPortConfigError", func() {
					Expect(IsInvalidPortConfig(err)).To(BeTrue())
					Expect(err.Error()).To(Equal(`Cannot parse app config. Multiple components export port '80/tcp' in pod 'ns5'.`))
				})

			})

			Describe("parsing invalid ports configs in pods, cannot duplicate ports in a single pod (mixed with/without protocol)", func() {
				var err error

				BeforeEach(func() {
					appConfig := testApp(
						testService("session1",
							addPorts(testComponent("alt1", "ns5"), generictypes.MustParseDockerPort("80")),
							addPorts(testComponent("alt2", "ns5"), generictypes.MustParseDockerPort("80/tcp")),
						),
					)

					err = appConfig.validate()
				})

				It("should throw error InvalidPortConfigError", func() {
					Expect(IsInvalidPortConfig(err)).To(BeTrue())
					Expect(err.Error()).To(Equal(`Cannot parse app config. Multiple components export port '80/tcp' in pod 'ns5'.`))
				})

			})
		*/
	})
})
