package userconfig

import (
	"encoding/json"
	"testing"

	"github.com/giantswarm/generic-types-go"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestUserConfigPodValidator(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "user config pod validator")
}

var _ = Describe("user config pod validator", func() {

	testComponent := func(name, pod string) ComponentConfig {
		return ComponentConfig{
			ComponentName: name,
			InstanceConfig: InstanceConfig{
				Image: generictypes.MustParseDockerImage("registry/namespace/repository:version"),
			},
			PodConfig: PodConfig{
				PodName: pod,
			},
		}
	}

	addVols := func(config ComponentConfig, vol ...VolumeConfig) ComponentConfig {
		if config.InstanceConfig.Volumes == nil {
			config.InstanceConfig.Volumes = vol
		} else {
			config.InstanceConfig.Volumes = append(config.InstanceConfig.Volumes, vol...)
		}
		return config
	}

	addDeps := func(config ComponentConfig, dep ...DependencyConfig) ComponentConfig {
		if config.InstanceConfig.Dependencies == nil {
			config.InstanceConfig.Dependencies = dep
		} else {
			config.InstanceConfig.Dependencies = append(config.InstanceConfig.Dependencies, dep...)
		}
		return config
	}

	testService := func(name string, components ...ComponentConfig) ServiceConfig {
		return ServiceConfig{
			ServiceName: name,
			Components:  components,
		}
	}

	testApp := func(services ...ServiceConfig) AppDefinition {
		return AppDefinition{
			AppName:  "test-app",
			Services: services,
		}
	}

	Describe("json.Unmarshal()", func() {
		Describe("parsing valid pods", func() {
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
		              "service_name": "session",
		              "components": [
		                {
		                  "component_name": "api",
		                  "image": "registry/namespace/repository:version",
		                  "pod": "a"
		                },
		                {
		                  "component_name": "redis",
		                  "image": "dockerfile/redis",
		                  "pod": "a"
		                },
		                {
		                  "component_name": "redis2",
		                  "image": "dockerfile/redis"
		                }
		              ]
		            },
		            {
		              "service_name": "session2",
		              "components": [
		                {
		                  "component_name": "api",
		                  "image": "registry/namespace/repository:version",
		                  "pod": "b"
		                },
		                {
		                  "component_name": "redis",
		                  "image": "dockerfile/redis",
		                  "pod": "b"
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

			It("should parse two services", func() {
				Expect(appConfig.Services).To(HaveLen(2))
			})

			It("should parse 3/2 components", func() {
				Expect(appConfig.Services[0].Components).To(HaveLen(3))
				Expect(appConfig.Services[1].Components).To(HaveLen(2))
			})

			It("should parse pod 'a' for 2 component, empty for third, 'b' for components in second service", func() {
				Expect(appConfig.Services[0].Components[0].PodName).To(Equal("a"))
				Expect(appConfig.Services[0].Components[1].PodName).To(Equal("a"))
				Expect(appConfig.Services[0].Components[2].PodName).To(Equal(""))
				Expect(appConfig.Services[1].Components[0].PodName).To(Equal("b"))
				Expect(appConfig.Services[1].Components[1].PodName).To(Equal("b"))
			})

		})

		Describe("parsing (invalid) cross service pod", func() {
			var err error

			BeforeEach(func() {
				appConfig := testApp(
					testService("session1",
						testComponent("api", "ns2"),
						testComponent("redis", "ns2"),
					),
					testService("session2",
						testComponent("api", "ns2"),
						testComponent("redis", "ns2"),
					),
				)

				err = appConfig.validate()
			})

			It("should throw error CrossServicePodError", func() {
				Expect(IsCrossServicePod(err)).To(BeTrue())
				Expect(err.Error()).To(Equal(`Cannot parse app config. Pod 'ns2' is used in multiple services.`))
			})

		})

		Describe("parsing (invalid) pods that are used only once", func() {
			var err error

			BeforeEach(func() {
				appConfig := testApp(
					testService("session1",
						testComponent("api", "ns3"),
						testComponent("redis", ""),
					),
				)

				err = appConfig.validate()
			})

			It("should throw error PodUsedOnlyOnceError", func() {
				Expect(IsPodUsedOnlyOnce(err)).To(BeTrue())
				Expect(err.Error()).To(Equal(`Cannot parse app config. Pod 'ns3' is used in only 1 component.`))
			})

		})

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

		Describe("parsing invalid volume configs in pods, invalid property combi path+volumes-from", func() {
			var err error

			BeforeEach(func() {
				appConfig := testApp(
					testService("session1",
						addVols(testComponent("api", "ns4"), VolumeConfig{Path: "/data1", Size: VolumeSize("27 GB"), VolumesFrom: "alt1"}),
						addVols(testComponent("alt1", "ns4"), VolumeConfig{VolumesFrom: "api"}),
					),
				)

				err = appConfig.validate()
			})

			It("should throw error InvalidVolumeConfigError", func() {
				Expect(IsInvalidVolumeConfig(err)).To(BeTrue())
				Expect(err.Error()).To(Equal(`Cannot parse volume config. Volumes-from for path '/data1' should be empty.`))
			})

		})

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

		Describe("parsing invalid dependency configs in pods, same name different sources", func() {
			var err error

			BeforeEach(func() {
				appConfig := testApp(
					testService("session1",
						addDeps(testComponent("alt1", "ns4"), DependencyConfig{Name: "redis1", Port: generictypes.MustParseDockerPort("6379")}),
						addDeps(testComponent("alt2", "ns4"), DependencyConfig{Name: "redis2", Port: generictypes.MustParseDockerPort("6379")}),
						testComponent("redis1", ""),
						testComponent("redis2", ""),
					),
				)

				err = appConfig.validate()
			})

			It("should throw error InvalidDependencyConfigError", func() {
				Expect(IsInvalidDependencyConfig(err)).To(BeTrue())
				Expect(err.Error()).To(Equal(`Cannot parse app config. Duplicate (but different) dependency with port '6379/tcp' in pod 'ns4'.`))
			})

		})

	})
})
