package userconfig_test

import (
	"encoding/json"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/giantswarm/user-config"
)

func TestUserConfigNamespaceValidator(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "user config namespace validator")
}

var _ = Describe("user config namespace validator", func() {

	Describe("json.Unmarshal()", func() {
		Describe("parsing valid namespaces", func() {
			var (
				err       error
				byteSlice []byte
				appConfig userconfig.AppDefinition
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
		                  "namespace": "a"
		                },
		                {
		                  "component_name": "redis",
		                  "image": "dockerfile/redis",
		                  "namespace": "a"
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
		                  "namespace": "b"
		                },
		                {
		                  "component_name": "redis",
		                  "image": "dockerfile/redis",
		                  "namespace": "b"
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
				Expect(appConfig.Services).To(HaveLen(2))
			})

			It("should parse 3/2 components", func() {
				Expect(appConfig.Services[0].Components).To(HaveLen(3))
				Expect(appConfig.Services[1].Components).To(HaveLen(2))
			})

			It("should parse namespace 'a' for 2 component, empty for third, 'b' for components in second service", func() {
				Expect(appConfig.Services[0].Components[0].NamespaceName).To(Equal("a"))
				Expect(appConfig.Services[0].Components[1].NamespaceName).To(Equal("a"))
				Expect(appConfig.Services[0].Components[2].NamespaceName).To(Equal(""))
				Expect(appConfig.Services[1].Components[0].NamespaceName).To(Equal("b"))
				Expect(appConfig.Services[1].Components[1].NamespaceName).To(Equal("b"))
			})

		})

		Describe("parsing (invalid) cross service namespaces", func() {
			var (
				err       error
				byteSlice []byte
				appConfig userconfig.AppDefinition
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
		                  "namespace": "ns2"
		                },
		                {
		                  "component_name": "redis",
		                  "image": "dockerfile/redis",
		                  "namespace": "ns2"
		                }
		              ]
		            },
		            {
		              "service_name": "session2",
		              "components": [
		                {
		                  "component_name": "api",
		                  "image": "registry/namespace/repository:version",
		                  "namespace": "ns2"
		                },
		                {
		                  "component_name": "redis",
		                  "image": "dockerfile/redis",
		                  "namespace": "ns2"
		                }
		              ]
		            }
		          ]
		        }`)

				err = json.Unmarshal(byteSlice, &appConfig)
			})

			It("should throw error CrossServiceNamespaceError", func() {
				Expect(userconfig.IsCrossServiceNamespace(err)).To(BeTrue())
				Expect(err.Error()).To(Equal(`Cannot parse app config. Namespace 'ns2' is used in multiple services.`))
			})

		})

		Describe("parsing (invalid) namespaces that are used only once", func() {
			var (
				err       error
				byteSlice []byte
				appConfig userconfig.AppDefinition
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
		                  "namespace": "ns3"
		                },
		                {
		                  "component_name": "redis",
		                  "image": "dockerfile/redis"
		                }
		              ]
		            }
		          ]
		        }`)

				err = json.Unmarshal(byteSlice, &appConfig)
			})

			It("should throw error NamespaceUsedOnlyOnceError", func() {
				Expect(userconfig.IsNamespaceUsedOnlyOnce(err)).To(BeTrue())
				Expect(err.Error()).To(Equal(`Cannot parse app config. Namespace 'ns3' is used in only 1 component.`))
			})

		})

		Describe("parsing valid volume configs in namespaces", func() {
			var (
				err       error
				byteSlice []byte
				appConfig userconfig.AppDefinition
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
		                  "namespace": "ns4",
		                  "volumes": [
		                  	{"path": "/data1", "size": "27 GB"}
		                  ]
		                },
		                {
		                  "component_name": "alt1",
		                  "image": "dockerfile/redis",
		                  "namespace": "ns4",
		                  "volumes": [
		                  	{"volumes-from": "api"}
		                  ]
		                },
		                {
		                  "component_name": "alt2",
		                  "image": "dockerfile/redis",
		                  "namespace": "ns4",
		                  "volumes": [
		                  	{"volume-from": "api", "volume-path": "/data1"}
		                  ]
		                },
		                {
		                  "component_name": "alt3",
		                  "image": "dockerfile/redis",
		                  "namespace": "ns4",
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

			It("should parse 2 components", func() {
				Expect(appConfig.Services[0].Components).To(HaveLen(4))
			})

			It("should parse one volume for each component", func() {
				Expect(appConfig.Services[0].Components[0].Volumes).To(HaveLen(1))
				Expect(appConfig.Services[0].Components[1].Volumes).To(HaveLen(1))
				Expect(appConfig.Services[0].Components[2].Volumes).To(HaveLen(1))
				Expect(appConfig.Services[0].Components[3].Volumes).To(HaveLen(1))

				Expect(appConfig.Services[0].Components[0].Volumes[0].Path).To(Equal("/data1"))
				Expect(appConfig.Services[0].Components[0].Volumes[0].Size).To(Equal(userconfig.VolumeSize("27 GB")))
				Expect(appConfig.Services[0].Components[1].Volumes[0].VolumesFrom).To(Equal("api"))
				Expect(appConfig.Services[0].Components[2].Volumes[0].VolumeFrom).To(Equal("api"))
				Expect(appConfig.Services[0].Components[2].Volumes[0].VolumePath).To(Equal("/data1"))
				Expect(appConfig.Services[0].Components[3].Volumes[0].VolumeFrom).To(Equal("api"))
				Expect(appConfig.Services[0].Components[3].Volumes[0].VolumePath).To(Equal("/data1"))
				Expect(appConfig.Services[0].Components[3].Volumes[0].Path).To(Equal("/alt4"))

			})

		})

		Describe("parsing invalid volume configs in namespaces #1", func() {
			var (
				err       error
				byteSlice []byte
				appConfig userconfig.AppDefinition
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
		                  "namespace": "ns4",
		                  "volumes": [
		                  	{"path": "/data1", "size": "27 GB", "volumes-from": "alt1"}
		                  ]
		                },
		                {
		                  "component_name": "alt1",
		                  "image": "dockerfile/redis",
		                  "namespace": "ns4",
		                  "volumes": [
		                  	{"volumes-from": "api"}
		                  ]
		                }
		              ]
		            }
		          ]
		        }`)

				err = json.Unmarshal(byteSlice, &appConfig)
			})

			It("should throw error InvalidVolumeConfigError", func() {
				Expect(userconfig.IsInvalidVolumeConfig(err)).To(BeTrue())
				Expect(err.Error()).To(Equal(`Cannot parse volume config. Volumes-from for path '/data1' should be empty.`))
			})

		})

		Describe("parsing invalid volume configs in namespaces #2", func() {
			var (
				err       error
				byteSlice []byte
				appConfig userconfig.AppDefinition
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
		                  "namespace": "ns4",
		                  "volumes": [
		                  	{"path": "/data1", "size": "27 GB", "volume-from": "alt1"}
		                  ]
		                },
		                {
		                  "component_name": "alt1",
		                  "image": "dockerfile/redis",
		                  "namespace": "ns4",
		                  "volumes": [
		                  	{"volumes-from": "api"}
		                  ]
		                }
		              ]
		            }
		          ]
		        }`)

				err = json.Unmarshal(byteSlice, &appConfig)
			})

			It("should throw error InvalidVolumeConfigError", func() {
				Expect(userconfig.IsInvalidVolumeConfig(err)).To(BeTrue())
				Expect(err.Error()).To(Equal(`Cannot parse volume config. Volume-from for path '/data1' should be empty.`))
			})

		})

		Describe("parsing invalid volume configs in namespaces #3", func() {
			var (
				err       error
				byteSlice []byte
				appConfig userconfig.AppDefinition
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
		                  "namespace": "ns4",
		                  "volumes": [
		                  	{"path": "/data1", "size": "27 GB", "volume-path": "/alt1"}
		                  ]
		                },
		                {
		                  "component_name": "alt1",
		                  "image": "dockerfile/redis",
		                  "namespace": "ns4",
		                  "volumes": [
		                  	{"volumes-from": "api"}
		                  ]
		                }
		              ]
		            }
		          ]
		        }`)

				err = json.Unmarshal(byteSlice, &appConfig)
			})

			It("should throw error InvalidVolumeConfigError", func() {
				Expect(userconfig.IsInvalidVolumeConfig(err)).To(BeTrue())
				Expect(err.Error()).To(Equal(`Cannot parse volume config. Volume-path for path '/data1' should be empty.`))
			})

		})

		Describe("parsing invalid volume configs in namespaces #4", func() {
			var (
				err       error
				byteSlice []byte
				appConfig userconfig.AppDefinition
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
		                  "namespace": "ns4",
		                  "volumes": [
		                  	{"path": "/data1", "size": "27 GB"}
		                  ]
		                },
		                {
		                  "component_name": "alt1",
		                  "image": "dockerfile/redis",
		                  "namespace": "ns4",
		                  "volumes": [
		                  	{"volumes-from": "api", "path": "/alt1"}
		                  ]
		                }
		              ]
		            }
		          ]
		        }`)

				err = json.Unmarshal(byteSlice, &appConfig)
			})

			It("should throw error InvalidVolumeConfigError", func() {
				Expect(userconfig.IsInvalidVolumeConfig(err)).To(BeTrue())
				Expect(err.Error()).To(Equal(`Cannot parse volume config. Path for volumes-from 'api' should be empty.`))
			})

		})

		Describe("parsing invalid volume configs in namespaces #5", func() {
			var (
				err       error
				byteSlice []byte
				appConfig userconfig.AppDefinition
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
		                  "namespace": "ns4",
		                  "volumes": [
		                  	{"path": "/data1", "size": "27 GB"}
		                  ]
		                },
		                {
		                  "component_name": "alt1",
		                  "image": "dockerfile/redis",
		                  "namespace": "ns4",
		                  "volumes": [
		                  	{"volumes-from": "api", "size": "5 GB"}
		                  ]
		                }
		              ]
		            }
		          ]
		        }`)

				err = json.Unmarshal(byteSlice, &appConfig)
			})

			It("should throw error InvalidVolumeConfigError", func() {
				Expect(userconfig.IsInvalidVolumeConfig(err)).To(BeTrue())
				Expect(err.Error()).To(Equal(`Cannot parse volume config. Size for volumes-from 'api' should be empty.`))
			})

		})

		Describe("parsing invalid volume configs in namespaces #5", func() {
			var (
				err       error
				byteSlice []byte
				appConfig userconfig.AppDefinition
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
		                  "namespace": "ns4",
		                  "volumes": [
		                  	{"path": "/data1", "size": "27 GB"}
		                  ]
		                },
		                {
		                  "component_name": "alt1",
		                  "image": "dockerfile/redis",
		                  "namespace": "ns4",
		                  "volumes": [
		                  	{"volumes-from": "api", "volume-from": "api"}
		                  ]
		                }
		              ]
		            }
		          ]
		        }`)

				err = json.Unmarshal(byteSlice, &appConfig)
			})

			It("should throw error InvalidVolumeConfigError", func() {
				Expect(userconfig.IsInvalidVolumeConfig(err)).To(BeTrue())
				Expect(err.Error()).To(Equal(`Cannot parse volume config. Volume-from for volumes-from 'api' should be empty.`))
			})

		})

		Describe("parsing invalid volume configs in namespaces #6", func() {
			var (
				err       error
				byteSlice []byte
				appConfig userconfig.AppDefinition
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
		                  "namespace": "ns4",
		                  "volumes": [
		                  	{"path": "/data1", "size": "27 GB"}
		                  ]
		                },
		                {
		                  "component_name": "alt1",
		                  "image": "dockerfile/redis",
		                  "namespace": "ns4",
		                  "volumes": [
		                  	{"volume-from": "api", "volume-path": "/data1", "size": "5GB"}
		                  ]
		                }
		              ]
		            }
		          ]
		        }`)

				err = json.Unmarshal(byteSlice, &appConfig)
			})

			It("should throw error InvalidVolumeConfigError", func() {
				Expect(userconfig.IsInvalidVolumeConfig(err)).To(BeTrue())
				Expect(err.Error()).To(Equal(`Cannot parse volume config. Size for volume-from 'api' should be empty.`))
			})

		})

		Describe("parsing invalid volume configs in namespaces #7", func() {
			var (
				err       error
				byteSlice []byte
				appConfig userconfig.AppDefinition
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
		                  "namespace": "ns4",
		                  "volumes": [
		                  	{"path": "/data1", "size": "27 GB"}
		                  ]
		                },
		                {
		                  "component_name": "alt1",
		                  "image": "dockerfile/redis",
		                  "namespace": "ns4",
		                  "volumes": [
		                  	{"volumes-from": "alt1"}
		                  ]
		                }
		              ]
		            }
		          ]
		        }`)

				err = json.Unmarshal(byteSlice, &appConfig)
			})

			It("should throw error InvalidVolumeConfigError", func() {
				Expect(userconfig.IsInvalidVolumeConfig(err)).To(BeTrue())
				Expect(err.Error()).To(Equal(`Cannot parse volume config. Cannot refer to own component 'alt1'.`))
			})

		})

		Describe("parsing invalid volume configs in namespaces #8", func() {
			var (
				err       error
				byteSlice []byte
				appConfig userconfig.AppDefinition
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
		                  "namespace": "ns4",
		                  "volumes": [
		                  	{"path": "/data1", "size": "27 GB"}
		                  ]
		                },
		                {
		                  "component_name": "alt1",
		                  "image": "dockerfile/redis",
		                  "namespace": "ns4",
		                  "volumes": [
		                  	{"volume-from": "alt1", "volume-path": "/data1"}
		                  ]
		                }
		              ]
		            }
		          ]
		        }`)

				err = json.Unmarshal(byteSlice, &appConfig)
			})

			It("should throw error InvalidVolumeConfigError", func() {
				Expect(userconfig.IsInvalidVolumeConfig(err)).To(BeTrue())
				Expect(err.Error()).To(Equal(`Cannot parse volume config. Cannot refer to own component 'alt1'.`))
			})

		})

		Describe("parsing invalid volume configs in namespaces #9", func() {
			var (
				err       error
				byteSlice []byte
				appConfig userconfig.AppDefinition
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
		                  "namespace": "ns4",
		                  "volumes": [
		                  	{"path": "/data1", "size": "27 GB"}
		                  ]
		                },
		                {
		                  "component_name": "alt1",
		                  "image": "dockerfile/redis",
		                  "namespace": "ns4",
		                  "volumes": [
		                  	{"volume-from": "api", "volume-path": "/unknown"}
		                  ]
		                }
		              ]
		            }
		          ]
		        }`)

				err = json.Unmarshal(byteSlice, &appConfig)
			})

			It("should throw error InvalidVolumeConfigError", func() {
				Expect(userconfig.IsInvalidVolumeConfig(err)).To(BeTrue())
				Expect(err.Error()).To(Equal(`Cannot parse volume config. Cannot find path '/unknown' on component 'api'.`))
			})

		})

		Describe("parsing invalid volume configs in namespaces #10", func() {
			var (
				err       error
				byteSlice []byte
				appConfig userconfig.AppDefinition
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
		                  "namespace": "ns4",
		                  "volumes": [
		                  	{"path": "/xdata1", "size": "27 GB"}
		                  ]
		                },
		                {
		                  "component_name": "alt1",
		                  "image": "dockerfile/redis",
		                  "namespace": "ns4",
		                  "volumes": [
		                  	{"volumes-from": "api"},
		                  	{"path": "/xdata1", "size": "5GB"}
		                  ]
		                }
		              ]
		            }
		          ]
		        }`)

				err = json.Unmarshal(byteSlice, &appConfig)
			})

			It("should throw error DuplicateVolumePathError", func() {
				Expect(userconfig.IsDuplicateVolumePath(err)).To(BeTrue())
				Expect(err.Error()).To(Equal(`Cannot parse app config. Duplicate volume '/xdata1' found in component 'alt1'.`))
			})

		})

		Describe("parsing invalid volume configs in namespaces #11", func() {
			var (
				err       error
				byteSlice []byte
				appConfig userconfig.AppDefinition
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
		                  "namespace": "ns4",
		                  "volumes": [
		                  	{"path": "/xdata1", "size": "27 GB"}
		                  ]
		                },
		                {
		                  "component_name": "alt1",
		                  "image": "dockerfile/redis",
		                  "namespace": "ns4",
		                  "volumes": [
		                  	{"path": "/xdata1", "size": "27 GB"},
		                  	{"volume-from": "api", "volume-path": "/xdata1"}
		                  ]
		                }
		              ]
		            }
		          ]
		        }`)

				err = json.Unmarshal(byteSlice, &appConfig)
			})

			It("should throw error DuplicateVolumePathError", func() {
				Expect(userconfig.IsDuplicateVolumePath(err)).To(BeTrue())
				Expect(err.Error()).To(Equal(`Cannot parse app config. Duplicate volume '/xdata1' found in component 'alt1'.`))
			})

		})

		Describe("parsing invalid volume configs in namespaces #12", func() {
			var (
				err       error
				byteSlice []byte
				appConfig userconfig.AppDefinition
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
		                  "namespace": "ns4",
		                  "volumes": [
		                  	{"path": "/xdata1", "size": "27 GB"}
		                  ]
		                },
		                {
		                  "component_name": "alt1",
		                  "image": "dockerfile/redis",
		                  "namespace": "ns4",
		                  "volumes": [
		                  	{"volumes-from": "api"}
		                  ]
		                },
		                {
		                  "component_name": "alt2",
		                  "image": "dockerfile/redis",
		                  "namespace": "ns4",
		                  "volumes": [
		                  	{"volumes-from": "api"},
		                  	{"volumes-from": "alt1"}
		                  ]
		                }
		              ]
		            }
		          ]
		        }`)

				err = json.Unmarshal(byteSlice, &appConfig)
			})

			It("should throw error DuplicateVolumePathError", func() {
				Expect(userconfig.IsDuplicateVolumePath(err)).To(BeTrue())
				Expect(err.Error()).To(Equal(`Cannot parse app config. Duplicate volume '/xdata1' found in component 'alt2'.`))
			})

		})

		Describe("parsing invalid volume configs in namespaces #13", func() {
			var (
				err       error
				byteSlice []byte
				appConfig userconfig.AppDefinition
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
		                  "namespace": "ns4",
		                  "volumes": [
		                  	{"volumes-from": "alt2"}
		                  ]
		                },
		                {
		                  "component_name": "alt1",
		                  "image": "dockerfile/redis",
		                  "namespace": "ns4",
		                  "volumes": [
		                  	{"volumes-from": "api"}
		                  ]
		                },
		                {
		                  "component_name": "alt2",
		                  "image": "dockerfile/redis",
		                  "namespace": "ns4",
		                  "volumes": [
		                  	{"volumes-from": "alt1"}
		                  ]
		                }
		              ]
		            }
		          ]
		        }`)

				err = json.Unmarshal(byteSlice, &appConfig)
			})

			It("should throw error InvalidVolumeConfigError", func() {
				Expect(userconfig.IsInvalidVolumeConfig(err)).To(BeTrue())
				Expect(err.Error()).To(Equal(`Cannot parse app config. Cycle in referenced components detected in 'alt2'.`))
			})

		})

	})
})
