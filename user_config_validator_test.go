package userconfig_test

import (
	"encoding/json"
	"fmt"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	userConfigPkg "github.com/giantswarm/user-config"
)

func TestUserConfigValidator(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "user config validator")
}

var _ = Describe("user config validator", func() {
	var (
		err       error
		byteSlice []byte
		appConfig userConfigPkg.AppConfig
	)

	BeforeEach(func() {
		err = nil
	})

	Describe("json.Unmarshal()", func() {
		Describe("parsing valid app-config", func() {
			BeforeEach(func() {
				byteSlice = []byte(`{
				          "app_name": "test-app-name",
				          "services": [
				            {
				              "service_name": "session",
				              "components": [
				                {
				                  "component_name": "api",
				                  "image": "registry/namespace/redis-example:0.0.2",
				                  "ports": [ "80/tcp" ],
				                  "dependencies": [
				                    { "name": "redis", "port": 6379, "same_machine": true }
				                  ],
				                  "domains": { "test.domain.io": "80" }
				                },
				                {
				                  "component_name": "redis",
				                  "image": "dockerfile/redis",
				                  "ports": [ "6379/tcp" ],
				                  "volumes": [
				                    { "path": "/data", "size": "5 GB" }
				                  ]
				                }
				              ]
				            }
				          ]
				        }`)

				appConfig = userConfigPkg.AppConfig{}
				err = json.Unmarshal(byteSlice, &appConfig)
				fmt.Printf("%#v\n", "#####################")
			})

			FIt("should not throw error", func() {
				Expect(err).To(BeNil())
			})

			//		It("should properly parse given app name", func() {
			//			Expect(appConfig.AppName).To(Equal("test-app-name"))
			//		})

			//		It("should parse one service", func() {
			//			Expect(appConfig.Services).To(HaveLen(1))
			//		})

			//		It("should parse one service", func() {
			//			Expect(appConfig.Services[0].ServiceName).To(Equal("session"))
			//		})

			//		It("should parse two components", func() {
			//			Expect(appConfig.Services[0].Components).To(HaveLen(2))
			//		})

			//		It("should parse one domain for component", func() {
			//			Expect(appConfig.Services[0].Components[0].Domains).To(HaveLen(1))
			//		})

			//		It("should parse correct component domain", func() {
			//			Expect(appConfig.Services[0].Components[0].Domains["test.domain.io"]).To(Equal("80"))
			//		})

			//		It("should parse correct component image 1", func() {
			//			Expect(appConfig.Services[0].Components[0].Image).NotTo(BeEmpty())
			//		})

			//		It("should parse correct component image 2", func() {
			//			Expect(appConfig.Services[0].Components[1].Image).NotTo(BeEmpty())
			//		})
		})

		//	Describe("parsing app-config with unknown fields", func() {
		//		BeforeEach(func() {
		//			byteSlice = []byte(`{
		//        "app_name": "test-app-name",
		//        "foo": 47,
		//        "services": [
		//          {
		//            "service_name": "session"
		//          }
		//        ]
		//      }`)

		//			appConfig = userConfigPkg.AppConfig{}
		//			err = json.Unmarshal(byteSlice, &appConfig)
		//		})

		//		It("should throw error ErrUnknownJSONField", func() {
		//			Expect(userConfigPkg.IsErrUnknownJsonField(err)).To(BeTrue())
		//		})

		//		It("should not parse given app name", func() {
		//			Expect(appConfig.AppName).To(Equal(""))
		//		})
		//	})
	})
})
