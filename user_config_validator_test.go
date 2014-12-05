package userconfig_test

import (
	"encoding/json"
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
                  "image": "registry/namespace/repository:version",
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
			})

			It("should not throw error", func() {
				Expect(err).To(BeNil())
			})

			It("should properly parse given app name", func() {
				Expect(appConfig.AppName).To(Equal("test-app-name"))
			})

			It("should parse one service", func() {
				Expect(appConfig.Services).To(HaveLen(1))
			})

			It("should parse one service", func() {
				Expect(appConfig.Services[0].ServiceName).To(Equal("session"))
			})

			It("should parse two components", func() {
				Expect(appConfig.Services[0].Components).To(HaveLen(2))
			})

			It("should parse one domain for component", func() {
				Expect(appConfig.Services[0].Components[0].Domains).To(HaveLen(1))
			})

			It("should parse correct component domain", func() {
				Expect(appConfig.Services[0].Components[0].Domains["test.domain.io"]).To(Equal("80"))
			})

			It("should parse correct component image 1", func() {
				Expect(appConfig.Services[0].Components[0].Image.Registry).To(Equal("registry"))
				Expect(appConfig.Services[0].Components[0].Image.Namespace).To(Equal("namespace"))
				Expect(appConfig.Services[0].Components[0].Image.Repository).To(Equal("repository"))
				Expect(appConfig.Services[0].Components[0].Image.Version).To(Equal("version"))
			})

			It("should parse correct component image 2", func() {
				Expect(appConfig.Services[0].Components[1].Image.Registry).To(Equal(""))
				Expect(appConfig.Services[0].Components[1].Image.Namespace).To(Equal("dockerfile"))
				Expect(appConfig.Services[0].Components[1].Image.Repository).To(Equal("redis"))
				Expect(appConfig.Services[0].Components[1].Image.Version).To(Equal(""))
			})
		})

		Describe("parsing app-config with missing fields", func() {
			BeforeEach(func() {
				byteSlice = []byte(`{
          "app_name": "test-app-name",
          "foo": 47,
          "services": [
            {
              "service_name": "session"
            }
          ]
        }`)

				appConfig = userConfigPkg.AppConfig{}
				err = json.Unmarshal(byteSlice, &appConfig)
			})

			It("should throw error ErrUnknownJSONField", func() {
				Expect(userConfigPkg.IsErrUnknownJsonField(err)).To(BeTrue())
				Expect(err.Error()).To(Equal(`Cannot parse app config. Unknown field '["foo"]' detected.`))
			})

			It("should not parse given app name", func() {
				Expect(appConfig.AppName).To(Equal(""))
			})
		})

		Describe("parsing app-config with unknown fields", func() {
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
                  "volume": [
                    { "path": "/data", "size": "5 GB" }
                  ]
                }
              ]
            }
          ]
        }`)

				appConfig = userConfigPkg.AppConfig{}
				err = json.Unmarshal(byteSlice, &appConfig)
			})

			It("should detect first occuring error and throw ErrUnknownJSONField", func() {
				Expect(userConfigPkg.IsErrUnknownJsonField(err)).To(BeTrue())
				Expect(err.Error()).To(Equal(`Cannot parse app config. Unknown field '["services"][0]["components"][1]["volume"]' detected.`))
			})

			It("should not parse given app name", func() {
				Expect(appConfig.AppName).To(Equal(""))
			})
		})

		Describe("fix app-config fields", func() {
			Describe("with actually valid field names", func() {
				BeforeEach(func() {
					byteSlice = []byte(`{
            "appName": "test-app-name",
            "Services": [
              {
                "service_name": "session",
                "Components": [
                  {
                    "component_name": "api",
                    "image": "registry/namespace/repository:version",
                    "ports": [ "80/tcp" ],
                    "dependencies": [
                      { "name": "redis", "port": 6379, "same_machine": true }
                    ],
                    "Domains": { "test.domain.io": "80" }
                  },
                  {
                    "ComponentName": "redis",
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
				})

				It("should not throw error", func() {
					Expect(err).To(BeNil())
				})

				It("should properly parse given app name", func() {
					Expect(appConfig.AppName).To(Equal("test-app-name"))
				})

				It("should parse one service", func() {
					Expect(appConfig.Services).To(HaveLen(1))
				})

				It("should parse one service", func() {
					Expect(appConfig.Services[0].ServiceName).To(Equal("session"))
				})

				It("should parse two components", func() {
					Expect(appConfig.Services[0].Components).To(HaveLen(2))
				})

				It("should parse one domain for component", func() {
					Expect(appConfig.Services[0].Components[0].Domains).To(HaveLen(1))
				})

				It("should parse correct component domain", func() {
					Expect(appConfig.Services[0].Components[0].Domains["test.domain.io"]).To(Equal("80"))
				})

				It("should parse correct component image 1", func() {
					Expect(appConfig.Services[0].Components[0].Image.Registry).To(Equal("registry"))
					Expect(appConfig.Services[0].Components[0].Image.Namespace).To(Equal("namespace"))
					Expect(appConfig.Services[0].Components[0].Image.Repository).To(Equal("repository"))
					Expect(appConfig.Services[0].Components[0].Image.Version).To(Equal("version"))
				})

				It("should parse correct component image 2", func() {
					Expect(appConfig.Services[0].Components[1].Image.Registry).To(Equal(""))
					Expect(appConfig.Services[0].Components[1].Image.Namespace).To(Equal("dockerfile"))
					Expect(appConfig.Services[0].Components[1].Image.Repository).To(Equal("redis"))
					Expect(appConfig.Services[0].Components[1].Image.Version).To(Equal(""))
				})
			})

			Describe("with actually invalid field names", func() {
				BeforeEach(func() {
					byteSlice = []byte(`{
            "app_name": "test-app-name",
            "services": [
              {
                "service_name": "session",
                "compOnents": [
                  {
                    "component_name": "api",
                    "image": "registry/namespace/repository:version",
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

					appConfig = userConfigPkg.AppConfig{
						// Don't fix the given JSON data's field names, but still validate
						// it as it is.
						IsUserData: true,
					}

					err = json.Unmarshal(byteSlice, &appConfig)
				})

				It("should throw error", func() {
					Expect(userConfigPkg.IsErrUnknownJsonField(err)).To(BeTrue())
					Expect(err.Error()).To(Equal(`Cannot parse app config. Unknown field '["services"][0]["compOnents"]' detected.`))
				})

				It("should not parse given app name", func() {
					Expect(appConfig.AppName).To(Equal(""))
				})
			})
		})
	})
})
