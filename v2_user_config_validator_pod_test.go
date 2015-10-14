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
	testComponent := func() *ComponentDefinition {
		return &ComponentDefinition{
			Image: MustParseImageDefinition("registry/namespace/repository:version"),
		}
	}

	setPod := func(config *ComponentDefinition, pod PodEnum) *ComponentDefinition {
		config.Pod = pod
		return config
	}

	addVols := func(config *ComponentDefinition, vol ...VolumeConfig) *ComponentDefinition {
		if config.Volumes == nil {
			config.Volumes = vol
		} else {
			config.Volumes = append(config.Volumes, vol...)
		}
		return config
	}

	addLinks := func(config *ComponentDefinition, link ...LinkDefinition) *ComponentDefinition {
		if config.Links == nil {
			config.Links = link
		} else {
			config.Links = append(config.Links, link...)
		}
		return config
	}

	addPorts := func(config *ComponentDefinition, ports ...generictypes.DockerPort) *ComponentDefinition {
		if config.Ports == nil {
			config.Ports = ports
		} else {
			config.Ports = append(config.Ports, ports...)
		}
		return config
	}

	addScale := func(config *ComponentDefinition, min, max int, placement Placement) *ComponentDefinition {
		config.Scale = &ScaleDefinition{
			Min:       min,
			Max:       max,
			Placement: placement,
		}
		return config
	}

	testApp := func() ComponentDefinitions {
		return make(ComponentDefinitions)
	}

	maxScale := 7
	validate := func(nds ComponentDefinitions) error {
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
		Describe("parsing (valid) pod==children on a component that has 2 children", func() {
			var err error

			BeforeEach(func() {
				components := testApp()
				components["component/a"] = setPod(testComponent(), PodChildren)
				components["component/a/b"] = testComponent()
				components["component/a/c"] = testComponent()

				err = validate(components)
			})

			It("should not throw any errors", func() {
				Expect(err).To(BeNil())
			})
		})

		Describe("parsing (invalid) pod==children on a component that has no children", func() {
			var err error

			BeforeEach(func() {
				components := testApp()
				components["component/a"] = setPod(testComponent(), PodChildren)

				err = validate(components)
			})

			It("should throw error IsInvalidPodConfig", func() {
				Expect(IsInvalidPodConfig(err)).To(BeTrue())
				Expect(err.Error()).To(Equal(`component 'component/a' must have at least 2 child components because if defines 'pod' as 'children'`))
			})
		})

		Describe("parsing (invalid) pod==children on a component that has 2 children (recursive), but only 1 direct", func() {
			var err error

			BeforeEach(func() {
				components := testApp()
				components["component/a"] = setPod(testComponent(), PodChildren)
				components["component/a/b"] = testComponent()
				components["component/a/b/c"] = testComponent()

				err = validate(components)
			})

			It("should throw error IsInvalidPodConfig", func() {
				Expect(IsInvalidPodConfig(err)).To(BeTrue())
				Expect(err.Error()).To(Equal(`component 'component/a' must have at least 2 child components because if defines 'pod' as 'children'`))
			})
		})

		Describe("parsing (valid) pod==inherit on a component that has 2 children (recursive)", func() {
			var err error

			BeforeEach(func() {
				components := testApp()
				components["component/a"] = setPod(testComponent(), PodInherit)
				components["component/a/b"] = testComponent()
				components["component/a/b/c"] = testComponent()

				err = validate(components)
			})

			It("should not throw any errors", func() {
				Expect(err).To(BeNil())
			})
		})

		Describe("parsing (invalid) pod==inherit on a component that has not enough children", func() {
			var err error

			BeforeEach(func() {
				components := testApp()
				components["component/a"] = setPod(testComponent(), PodInherit)
				components["component/a/b"] = testComponent()

				err = validate(components)
			})

			It("should throw error IsInvalidPodConfig", func() {
				Expect(IsInvalidPodConfig(err)).To(BeTrue())
				Expect(err.Error()).To(Equal(`component 'component/a' must have at least 2 child components because if defines 'pod' as 'inherit'`))
			})
		})

		Describe("parsing (invalid) cannot specify pod value other than 'none' inside a pod", func() {
			var err error

			BeforeEach(func() {
				components := testApp()
				components["component/a"] = setPod(testComponent(), PodChildren)
				components["component/a/b"] = setPod(testComponent(), PodChildren)
				components["component/a/b/c1"] = testComponent()
				components["component/a/b/c2"] = testComponent()
				components["component/a/c"] = testComponent()

				err = validate(components)
			})

			It("should throw error IsInvalidPodConfig", func() {
				Expect(IsInvalidPodConfig(err)).To(BeTrue())
				Expect(err.Error()).To(Equal(`component 'component/a/b' must cannot set 'pod' to 'children' because it is already part of another pod`))
			})
		})

		Describe("parsing valid volume configs in pods", func() {
			var (
				err    error
				appDef V2AppDefinition
			)

			BeforeEach(func() {
				byteSlice := []byte(`{
		          "components": {
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

			It("should parse 5 components", func() {
				Expect(appDef.Components).To(HaveLen(5))
			})

			It("should parse one volume for each component", func() {
				Expect(appDef.Components[ComponentName("session1/api")].Volumes).To(HaveLen(1))
				Expect(appDef.Components[ComponentName("session1/alt1")].Volumes).To(HaveLen(1))
				Expect(appDef.Components[ComponentName("session1/alt2")].Volumes).To(HaveLen(1))
				Expect(appDef.Components[ComponentName("session1/alt3")].Volumes).To(HaveLen(1))

				Expect(appDef.Components[ComponentName("session1/api")].Volumes[0].Path).To(Equal("/data1"))
				Expect(appDef.Components[ComponentName("session1/api")].Volumes[0].Size).To(Equal(VolumeSize("27 GB")))
				Expect(appDef.Components[ComponentName("session1/alt1")].Volumes[0].VolumesFrom).To(Equal("session1/api"))
				Expect(appDef.Components[ComponentName("session1/alt2")].Volumes[0].VolumeFrom).To(Equal("session1/api"))
				Expect(appDef.Components[ComponentName("session1/alt2")].Volumes[0].VolumePath).To(Equal("/data1"))
				Expect(appDef.Components[ComponentName("session1/alt3")].Volumes[0].VolumeFrom).To(Equal("session1/api"))
				Expect(appDef.Components[ComponentName("session1/alt3")].Volumes[0].VolumePath).To(Equal("/data1"))
				Expect(appDef.Components[ComponentName("session1/alt3")].Volumes[0].Path).To(Equal("/alt4"))

			})
		})

		Describe("parsing invalid volume configs in pods, invalid property combi path+volumes-from", func() {
			var err error

			BeforeEach(func() {
				components := testApp()
				components["component/a"] = setPod(testComponent(), PodChildren)
				components["component/a/b1"] = addVols(testComponent(), VolumeConfig{Path: "/data1", Size: VolumeSize("27 GB"), VolumesFrom: "component/a/b2"})
				components["component/a/b2"] = addVols(testComponent(), VolumeConfig{VolumesFrom: "component/a/b1"})

				err = validate(components)
			})

			It("should throw error InvalidVolumeConfigError", func() {
				Expect(IsInvalidVolumeConfig(err)).To(BeTrue())
				Expect(err.Error()).To(Equal(`volumes-from for path '/data1' should be empty`))
			})
		})

		Describe("parsing invalid volume configs in pods, invalid property combi path+volume-from", func() {
			var err error

			BeforeEach(func() {
				components := testApp()
				components["component/a"] = setPod(testComponent(), PodChildren)
				components["component/a/b1"] = addVols(testComponent(), VolumeConfig{Path: "/data1", Size: VolumeSize("27 GB"), VolumeFrom: "component/a/b2"})
				components["component/a/b2"] = addVols(testComponent(), VolumeConfig{VolumesFrom: "component/a/b1"})

				err = validate(components)
			})

			It("should throw error InvalidVolumeConfigError", func() {
				Expect(IsInvalidVolumeConfig(err)).To(BeTrue())
				Expect(err.Error()).To(Equal(`volume-from for path '/data1' should be empty`))
			})
		})

		Describe("parsing invalid volume configs in pods, invalid property combi volumes-from+path", func() {
			var err error

			BeforeEach(func() {
				components := testApp()
				components["component/a"] = setPod(testComponent(), PodChildren)
				components["component/a/b1"] = addVols(testComponent(), VolumeConfig{Path: "/data1", Size: VolumeSize("27 GB")})
				components["component/a/b2"] = addVols(testComponent(), VolumeConfig{VolumesFrom: "component/a/b1", Path: "/alt1"})

				err = validate(components)
			})

			It("should throw error InvalidVolumeConfigError", func() {
				Expect(IsInvalidVolumeConfig(err)).To(BeTrue())
				Expect(err.Error()).To(Equal(`path for volumes-from 'component/a/b1' should be empty`))
			})
		})

		Describe("parsing invalid volume configs in pods, invalid property combi volumes-from+size", func() {
			var err error

			BeforeEach(func() {
				components := testApp()
				components["component/a"] = setPod(testComponent(), PodChildren)
				components["component/a/b1"] = addVols(testComponent(), VolumeConfig{Path: "/data1", Size: VolumeSize("27 GB")})
				components["component/a/b2"] = addVols(testComponent(), VolumeConfig{VolumesFrom: "component/a/b1", Size: VolumeSize("5 GB")})

				err = validate(components)
			})

			It("should throw error InvalidVolumeConfigError", func() {
				Expect(IsInvalidVolumeConfig(err)).To(BeTrue())
				Expect(err.Error()).To(Equal(`size for volumes-from 'component/a/b1' should be empty`))
			})
		})

		Describe("parsing invalid volume configs in pods, invalid property combi volumes-from+volume-from", func() {
			var err error

			BeforeEach(func() {
				components := testApp()
				components["component/a"] = setPod(testComponent(), PodChildren)
				components["component/a/b1"] = addVols(testComponent(), VolumeConfig{Path: "/data1", Size: VolumeSize("27 GB")})
				components["component/a/b2"] = addVols(testComponent(), VolumeConfig{VolumesFrom: "component/a/b1", VolumeFrom: "component/a/b1"})

				err = validate(components)
			})

			It("should throw error InvalidVolumeConfigError", func() {
				Expect(IsInvalidVolumeConfig(err)).To(BeTrue())
				Expect(err.Error()).To(Equal(`volume-from for volumes-from 'component/a/b1' should be empty`))
			})
		})

		Describe("parsing invalid volume configs in pods, invalid property combi volume-from+size", func() {
			var err error

			BeforeEach(func() {
				components := testApp()
				components["component/a"] = setPod(testComponent(), PodChildren)
				components["component/a/b1"] = addVols(testComponent(), VolumeConfig{Path: "/data1", Size: VolumeSize("27 GB")})
				components["component/a/b2"] = addVols(testComponent(), VolumeConfig{VolumeFrom: "component/a/b1", VolumePath: "/data1", Size: VolumeSize("5GB")})

				err = validate(components)
			})

			It("should throw error InvalidVolumeConfigError", func() {
				Expect(IsInvalidVolumeConfig(err)).To(BeTrue())
				Expect(err.Error()).To(Equal(`size for volume-from 'component/a/b1' should be empty`))
			})
		})

		Describe("parsing invalid volume configs in pods, volumes-from cannot refer to self", func() {
			var err error

			BeforeEach(func() {
				components := testApp()
				components["component/a"] = setPod(testComponent(), PodChildren)
				components["component/a/b1"] = addVols(testComponent(), VolumeConfig{Path: "/data1", Size: VolumeSize("27 GB")})
				components["component/a/b2"] = addVols(testComponent(), VolumeConfig{VolumesFrom: "component/a/b2"})

				err = validate(components)
			})

			It("should throw error InvalidVolumeConfigError", func() {
				Expect(IsInvalidVolumeConfig(err)).To(BeTrue())
				Expect(err.Error()).To(Equal(`cannot refer to own component 'component/a/b2'`))
			})
		})

		Describe("parsing invalid volume configs in pod, volume-from cannot refer to self", func() {
			var err error

			BeforeEach(func() {
				components := testApp()
				components["component/a"] = setPod(testComponent(), PodChildren)
				components["component/a/b1"] = addVols(testComponent(), VolumeConfig{Path: "/data1", Size: VolumeSize("27 GB")})
				components["component/a/b2"] = addVols(testComponent(), VolumeConfig{VolumeFrom: "component/a/b2", VolumePath: "/data1"})

				err = validate(components)
			})

			It("should throw error InvalidVolumeConfigError", func() {
				Expect(IsInvalidVolumeConfig(err)).To(BeTrue())
				Expect(err.Error()).To(Equal(`cannot refer to own component 'component/a/b2'`))
			})
		})

		Describe("parsing invalid volume configs in pods, unknown path in volume-path", func() {
			var err error

			BeforeEach(func() {
				components := testApp()
				components["component/a"] = setPod(testComponent(), PodChildren)
				components["component/a/b1"] = addVols(testComponent(), VolumeConfig{Path: "/data1", Size: VolumeSize("27 GB")})
				components["component/a/b2"] = addVols(testComponent(), VolumeConfig{VolumeFrom: "component/a/b1", VolumePath: "/unknown"})

				err = validate(components)
			})

			It("should throw error InvalidVolumeConfigError", func() {
				Expect(IsInvalidVolumeConfig(err)).To(BeTrue())
				Expect(err.Error()).To(Equal(`cannot find path '/unknown' on component 'component/a/b1'`))
			})
		})

		Describe("parsing invalid volume configs, duplicate volume via different postfixes", func() {
			var err error

			BeforeEach(func() {
				components := testApp()
				components["component/a"] = addVols(testComponent(),
					VolumeConfig{Path: "/xdata1", Size: VolumeSize("27 GB")},
					VolumeConfig{Path: "/xdata1/", Size: VolumeSize("27 GB")},
				)

				err = validate(components)
			})

			It("should throw error DuplicateVolumePathError", func() {
				Expect(IsDuplicateVolumePath(err)).To(BeTrue())
				Expect(err.Error()).To(Equal(`duplicate volume '/xdata1' found in component 'component/a'`))
			})
		})

		Describe("parsing invalid volume configs in pods, duplicate volume via volumes-from", func() {
			var err error

			BeforeEach(func() {
				components := testApp()
				components["component/a"] = setPod(testComponent(), PodChildren)
				components["component/a/b1"] = addVols(testComponent(), VolumeConfig{Path: "/xdata1", Size: VolumeSize("27 GB")})
				components["component/a/b2"] = addVols(testComponent(),
					VolumeConfig{VolumesFrom: "component/a/b1"},
					VolumeConfig{Path: "/xdata1", Size: VolumeSize("5GB")},
				)

				err = validate(components)
			})

			It("should throw error DuplicateVolumePathError", func() {
				//Expect(IsDuplicateVolumePath(err)).To(BeTrue())
				Expect(err.Error()).To(Equal(`duplicate volume '/xdata1' found in component 'component/a/b2'`))
			})
		})

		Describe("parsing invalid volume configs in pods, duplicate volume via volume-from", func() {
			var err error

			BeforeEach(func() {
				components := testApp()
				components["component/a"] = setPod(testComponent(), PodChildren)
				components["component/a/b1"] = addVols(testComponent(), VolumeConfig{Path: "/xdata1", Size: VolumeSize("27 GB")})
				components["component/a/b2"] = addVols(testComponent(),
					VolumeConfig{Path: "/xdata1", Size: VolumeSize("5GB")},
					VolumeConfig{VolumeFrom: "component/a/b1", VolumePath: "/xdata1"},
				)

				err = validate(components)
			})

			It("should throw error DuplicateVolumePathError", func() {
				Expect(IsDuplicateVolumePath(err)).To(BeTrue())
				Expect(err.Error()).To(Equal(`duplicate volume '/xdata1' found in component 'component/a/b2'`))
			})
		})

		Describe("parsing invalid volume configs in pods, duplicate volume via linked volumes-from", func() {
			var err error

			BeforeEach(func() {
				components := testApp()
				components["component/a"] = setPod(testComponent(), PodChildren)
				components["component/a/b1"] = addVols(testComponent(), VolumeConfig{Path: "/xdata1", Size: VolumeSize("27 GB")})
				components["component/a/b2"] = addVols(testComponent(), VolumeConfig{VolumesFrom: "component/a/b1"})
				components["component/a/b3"] = addVols(testComponent(),
					VolumeConfig{VolumesFrom: "component/a/b1"},
					VolumeConfig{VolumesFrom: "component/a/b2"},
				)

				err = validate(components)
			})

			It("should throw error DuplicateVolumePathError", func() {
				Expect(IsDuplicateVolumePath(err)).To(BeTrue())
				Expect(err.Error()).To(Equal(`duplicate volume '/xdata1' found in component 'component/a/b3'`))
			})
		})

		Describe("parsing invalid volume configs in pods, cycle in volumes-from references", func() {
			var err error

			BeforeEach(func() {
				components := testApp()
				components["component/a"] = setPod(testComponent(), PodChildren)
				components["component/a/b1"] = addVols(testComponent(), VolumeConfig{VolumesFrom: "component/a/b3"})
				components["component/a/b2"] = addVols(testComponent(), VolumeConfig{VolumesFrom: "component/a/b1"})
				components["component/a/b3"] = addVols(testComponent(), VolumeConfig{VolumesFrom: "component/a/b2"})

				err = validate(components)
			})

			It("should throw error VolumeCycleError", func() {
				Expect(IsVolumeCycle(err)).To(BeTrue())
			})
		})

		Describe("parsing dependency configs in pods, same name should result in same port", func() {
			var err error

			BeforeEach(func() {
				components := testApp()
				components["component/a"] = setPod(testComponent(), PodChildren)
				components["component/a/b1"] = addLinks(testComponent(), LinkDefinition{Component: "redis", TargetPort: generictypes.MustParseDockerPort("6379")})
				components["component/a/b2"] = addLinks(testComponent(), LinkDefinition{Component: "redis", TargetPort: generictypes.MustParseDockerPort("6379")})
				components["redis"] = addPorts(testComponent(), generictypes.MustParseDockerPort("6379"))

				err = validate(components)
			})

			It("should not throw an error", func() {
				Expect(err).To(BeNil())
			})
		})

		Describe("parsing invalid dependency configs in pods, same name different ports", func() {
			var err error

			BeforeEach(func() {
				components := testApp()
				components["component/a"] = setPod(testComponent(), PodChildren)
				components["component/a/b1"] = addLinks(testComponent(), LinkDefinition{Component: "redis1", TargetPort: generictypes.MustParseDockerPort("6379")})
				components["component/a/b2"] = addLinks(testComponent(), LinkDefinition{Component: "redis1", TargetPort: generictypes.MustParseDockerPort("1234")})
				components["redis1"] = addPorts(testComponent(), generictypes.MustParseDockerPort("6379"), generictypes.MustParseDockerPort("1234"))
				components["redis2"] = addPorts(testComponent(), generictypes.MustParseDockerPort("6379"))

				err = validate(components)
			})

			It("should throw error InvalidDependencyConfigError", func() {
				Expect(IsInvalidDependencyConfig(err)).To(BeTrue())
				Expect(err.Error()).To(Equal(`duplicate (with different ports) dependency 'redis1' in pod under 'component/a'`))
			})
		})

		Describe("parsing invalid dependency configs in pods, same alias different names", func() {
			var err error

			BeforeEach(func() {
				components := testApp()
				components["component/a"] = setPod(testComponent(), PodChildren)
				components["component/a/b1"] = addLinks(testComponent(), LinkDefinition{Alias: "db", Component: "redis1", TargetPort: generictypes.MustParseDockerPort("6379")})
				components["component/a/b2"] = addLinks(testComponent(), LinkDefinition{Alias: "db", Component: "redis2", TargetPort: generictypes.MustParseDockerPort("6379")})
				components["redis1"] = addPorts(testComponent(), generictypes.MustParseDockerPort("6379"))
				components["redis2"] = addPorts(testComponent(), generictypes.MustParseDockerPort("6379"))

				err = validate(components)
			})

			It("should throw error InvalidDependencyConfigError", func() {
				Expect(IsInvalidDependencyConfig(err)).To(BeTrue())
				Expect(err.Error()).To(Equal(`duplicate (with different names) dependency 'db' in pod under 'component/a'`))
			})
		})

		Describe("parsing invalid dependency configs in pods, one with aliases, other with (conflicting) name", func() {
			var err error

			BeforeEach(func() {
				components := testApp()
				components["component/a"] = setPod(testComponent(), PodChildren)
				components["component/a/b1"] = addLinks(testComponent(), LinkDefinition{Component: "component/redis", TargetPort: generictypes.MustParseDockerPort("6379")})
				components["component/a/b2"] = addLinks(testComponent(), LinkDefinition{Alias: "redis", Component: "component/redis2", TargetPort: generictypes.MustParseDockerPort("6379")})
				components["redis1"] = addPorts(testComponent(), generictypes.MustParseDockerPort("6379"))
				components["redis2"] = addPorts(testComponent(), generictypes.MustParseDockerPort("6379"))
				components["component/redis"] = addPorts(testComponent(), generictypes.MustParseDockerPort("6379"))
				components["component/redis2"] = addPorts(testComponent(), generictypes.MustParseDockerPort("6379"))

				err = validate(components)
			})

			It("should throw error InvalidDependencyConfigError", func() {
				Expect(IsInvalidDependencyConfig(err)).To(BeTrue())
				Expect(err.Error()).To(Equal(`duplicate (with different names) dependency 'redis' in pod under 'component/a'`))
			})
		})

		Describe("parsing invalid dependency configs in pods, same alias different ports", func() {
			var err error

			BeforeEach(func() {
				components := testApp()
				components["component/a"] = setPod(testComponent(), PodChildren)
				components["component/a/b1"] = addLinks(testComponent(), LinkDefinition{Alias: "db", Component: "redis", TargetPort: generictypes.MustParseDockerPort("6379")})
				components["component/a/b2"] = addLinks(testComponent(), LinkDefinition{Alias: "db", Component: "redis", TargetPort: generictypes.MustParseDockerPort("9736")})
				components["redis"] = addPorts(testComponent(), generictypes.MustParseDockerPort("6379"), generictypes.MustParseDockerPort("9736"))
				components["redis2"] = addPorts(testComponent(), generictypes.MustParseDockerPort("6379"))

				err = validate(components)
			})

			It("should throw error InvalidDependencyConfigError", func() {
				Expect(IsInvalidDependencyConfig(err)).To(BeTrue())
				Expect(err.Error()).To(Equal(`duplicate (with different ports) dependency 'db' in pod under 'component/a'`))
			})
		})

		Describe("parsing invalid dependency configs, linking to a component that is not a parent/sibling", func() {
			var err error

			BeforeEach(func() {
				components := testApp()
				components["component/a"] = testComponent()
				components["component/a/b1"] = addLinks(testComponent(), LinkDefinition{Component: "component/b/redis", TargetPort: generictypes.MustParseDockerPort("6379")})
				components["component/b/redis"] = addPorts(testComponent(), generictypes.MustParseDockerPort("6379"))

				err = validate(components)
			})

			It("should throw error InvalidLinkDefinitionError", func() {
				Expect(IsInvalidLinkDefinition(err)).To(BeTrue())
				Expect(err.Error()).To(Equal(`invalid link to component 'component/b/redis': component 'component/a/b1' is not allowed to link to it`))
			})
		})

		Describe("parsing valid dependency configs, linking to a component that is a parent/sibling", func() {
			var err error

			BeforeEach(func() {
				components := testApp()
				components["component/a"] = testComponent()
				components["component/a/b1"] = addLinks(testComponent(), LinkDefinition{Component: "component/a/redis", TargetPort: generictypes.MustParseDockerPort("6379")})
				components["component/a/redis"] = addPorts(testComponent(), generictypes.MustParseDockerPort("6379"))

				err = validate(components)
			})

			It("should now throw an error", func() {
				Expect(err).To(BeNil())
			})
		})

		Describe("parsing valid dependency configs, linking to a component that is a child", func() {
			var err error

			BeforeEach(func() {
				components := testApp()
				components["component/a"] = testComponent()
				components["component/a/b1"] = addLinks(testComponent(), LinkDefinition{Component: "component/a/b1/redis", TargetPort: generictypes.MustParseDockerPort("6379")})
				components["component/a/b1/redis"] = addPorts(testComponent(), generictypes.MustParseDockerPort("6379"))

				err = validate(components)
			})

			It("should now throw an error", func() {
				Expect(err).To(BeNil())
			})
		})

		Describe("parsing valid dependency configs, linking to a component that is a grandchild", func() {
			var err error

			BeforeEach(func() {
				components := testApp()
				components["component/a"] = testComponent()
				components["component/a/b1"] = addLinks(testComponent(), LinkDefinition{Component: "component/a/b1/c/redis", TargetPort: generictypes.MustParseDockerPort("6379")})
				components["component/a/b1/c/redis"] = addPorts(testComponent(), generictypes.MustParseDockerPort("6379"))

				err = validate(components)
			})

			It("should now throw an error", func() {
				Expect(err).To(BeNil())
			})
		})

		Describe("parsing valid scaling configs in pods, scaling values should be the same in all components of a pod or not set", func() {
			var err error

			BeforeEach(func() {
				components := testApp()
				components["component/a"] = setPod(testComponent(), PodChildren)
				components["component/a/b1"] = addScale(testComponent(), 1, maxScale, "")
				components["component/a/b2"] = addScale(testComponent(), 0, maxScale, "")
				components["component/a/b3"] = addScale(testComponent(), 1, 0, "")
				components["component/a/b4"] = testComponent()

				err = validate(components)
			})

			It("should not throw an error", func() {
				Expect(err).To(BeNil())
			})
		})

		Describe("parsing invalid scaling definition in pods, minimum scaling values should be the same in all components of a pod", func() {
			var err error

			BeforeEach(func() {
				components := testApp()
				components["component/a"] = setPod(testComponent(), PodChildren)
				components["component/a/b1"] = addScale(testComponent(), 1, 5, "")
				components["component/a/b2"] = addScale(testComponent(), 2, 5, "")

				err = validate(components)
			})

			It("should throw error InvalidScalingConfigError", func() {
				Expect(IsInvalidScalingConfig(err)).To(BeTrue())
				Expect(err.Error()).To(Equal(`different minimum scaling policies in pod under 'component/a'`))
			})
		})

		Describe("parsing invalid scaling definition in pods, maximum scaling values should be the same in all components of a pod", func() {
			var err error

			BeforeEach(func() {
				components := testApp()
				components["component/a"] = setPod(testComponent(), PodChildren)
				components["component/a/b1"] = addScale(testComponent(), 2, 5, "")
				components["component/a/b2"] = addScale(testComponent(), 2, 7, "")

				err = validate(components)
			})

			It("should throw error InvalidScalingConfigError", func() {
				Expect(IsInvalidScalingConfig(err)).To(BeTrue())
				Expect(err.Error()).To(Equal(`different maximum scaling policies in pod under 'component/a'`))
			})
		})

		Describe("parsing invalid scaling definition in pods, scaling placement values should be the same in all components of a pod", func() {
			var err error

			BeforeEach(func() {
				components := testApp()
				components["component/a"] = setPod(testComponent(), PodChildren)
				components["component/a/b1"] = addScale(testComponent(), 2, 5, DefaultPlacement)
				components["component/a/b2"] = addScale(testComponent(), 2, 5, OnePerMachinePlacement)

				err = validate(components)
			})

			It("should throw error InvalidScalingConfigError", func() {
				Expect(IsInvalidScalingConfig(err)).To(BeTrue())
				Expect(err.Error()).To(Equal(`different scaling placement policies in pod under 'component/a'`))
			})
		})

		Describe("parsing valid scaling definition in pods, maximum scaling values should be the same or not set in all components of a pod", func() {
			var err error

			BeforeEach(func() {
				components := testApp()
				components["component/a"] = setPod(testComponent(), PodChildren)
				components["component/a/b1"] = testComponent()                     // Scale not set here
				components["component/a/b2"] = addScale(testComponent(), 2, 7, "") // Scale set here

				err = validate(components)
			})

			It("should not throw an error", func() {
				Expect(err).To(BeNil())
			})
		})

		Describe("parsing valid scaling definition in pods, scaling values should be the same or not set in all components of a pod", func() {
			var err error

			BeforeEach(func() {
				components := testApp()
				components["component/a"] = setPod(testComponent(), PodChildren)
				components["component/a/b1"] = addScale(testComponent(), 0, 0, OnePerMachinePlacement) // Scale placement set here
				components["component/a/b2"] = addScale(testComponent(), 2, 7, "")                     // Scale min, max set here

				err = validate(components)
			})

			It("should not throw an error", func() {
				Expect(err).To(BeNil())
			})
		})

		Describe("parsing invalid ports configs in pods, cannot duplicate ports in a single pod", func() {
			var err error

			BeforeEach(func() {
				components := testApp()
				components["component/a"] = setPod(testComponent(), PodChildren)
				components["component/a/b1"] = addPorts(testComponent(), generictypes.MustParseDockerPort("80"))
				components["component/a/b2"] = addPorts(testComponent(), generictypes.MustParseDockerPort("80"))

				err = validate(components)
			})

			It("should throw error InvalidPortConfigError", func() {
				Expect(IsInvalidPortConfig(err)).To(BeTrue())
				Expect(err.Error()).To(Equal(`multiple components export port '80/tcp' in pod under 'component/a'`))
			})
		})

		Describe("parsing invalid ports configs in pods, cannot duplicate ports in a single pod (mixed with/without protocol)", func() {
			var err error

			BeforeEach(func() {
				components := testApp()
				components["component/a"] = setPod(testComponent(), PodChildren)
				components["component/a/b1"] = addPorts(testComponent(), generictypes.MustParseDockerPort("80"))
				components["component/a/b2"] = addPorts(testComponent(), generictypes.MustParseDockerPort("80/tcp"))

				err = validate(components)
			})

			It("should throw error InvalidPortConfigError", func() {
				Expect(IsInvalidPortConfig(err)).To(BeTrue())
				Expect(err.Error()).To(Equal(`multiple components export port '80/tcp' in pod under 'component/a'`))
			})
		})

		Describe("parsing links to the same component with different ports", func() {
			var err error

			BeforeEach(func() {
				components := testApp()
				components["component/a"] = addLinks(testComponent(),
					LinkDefinition{Component: "redisX", TargetPort: generictypes.MustParseDockerPort("6379")},
					LinkDefinition{Component: "redisX", TargetPort: generictypes.MustParseDockerPort("1234")})
				components["redisX"] = addPorts(testComponent(), generictypes.MustParseDockerPort("6379"), generictypes.MustParseDockerPort("1234"))

				err = validate(components)
			})

			It("should NOT throw error", func() {
				Expect(err).To(BeNil())
			})
		})

		Describe("parsing links to a component within the same pod having duplicated names", func() {
			var err error

			BeforeEach(func() {
				components := testApp()
				components["component"] = setPod(testComponent(), PodChildren)
				components["component/a"] = addLinks(testComponent(),
					LinkDefinition{Component: "component/b", TargetPort: generictypes.MustParseDockerPort("6379")},
					LinkDefinition{Component: "component/b", TargetPort: generictypes.MustParseDockerPort("1234")})
				components["component/b"] = addPorts(testComponent(), generictypes.MustParseDockerPort("6379"), generictypes.MustParseDockerPort("1234"))

				err = validate(components)
			})

			It("should throw error InvalidLinkDefinitionError", func() {
				Expect(IsInvalidDependencyConfig(err)).To(BeTrue())
				Expect(err.Error()).To(Equal(`duplicate (with different ports) dependency 'b' in pod under 'component'`))
			})
		})

		Describe("parsing links to a component within the same pod having duplicated names, but alias set", func() {
			var err error

			BeforeEach(func() {
				components := testApp()
				components["component"] = setPod(testComponent(), PodChildren)
				components["component/a"] = addLinks(testComponent(),
					LinkDefinition{Component: "component/b", TargetPort: generictypes.MustParseDockerPort("6379")},
					LinkDefinition{Component: "component/b", TargetPort: generictypes.MustParseDockerPort("1234"), Alias: "foo"}) // has alias that prevents conflict
				components["component/b"] = addPorts(testComponent(), generictypes.MustParseDockerPort("6379"), generictypes.MustParseDockerPort("1234"))

				err = validate(components)
			})

			It("should NOT throw error", func() {
				Expect(err).To(BeNil())
			})
		})
	})
})
