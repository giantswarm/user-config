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

			It("should throw error ErrCrossServiceNamespace", func() {
				Expect(userconfig.IsErrCrossServiceNamespace(err)).To(BeTrue())
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

			It("should throw error ErrNamespaceUsedOnlyOnce", func() {
				Expect(userconfig.IsErrNamespaceUsedOnlyOnce(err)).To(BeTrue())
				Expect(err.Error()).To(Equal(`Cannot parse app config. Namespace 'ns3' is used in only 1 component.`))
			})

		})

	})
})
