package userconfig_test

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/giantswarm/generic-types-go"
	"github.com/giantswarm/user-config"
)

func TestMarshalUnmarshal(t *testing.T) {
	app := ExampleDefinition()

	data, err := json.Marshal(app)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var app2 userconfig.AppDefinition
	if err := json.Unmarshal(data, &app2); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if !reflect.DeepEqual(app, app2) {
		t.Fatalf("Objects differ:\n%v\n%v", app, app2)
	}
}

func TestUserConfigValidator(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "user config validator")
}

var _ = Describe("user config validator", func() {
	var (
		err       error
		byteSlice []byte
		appConfig userconfig.AppDefinition
	)

	BeforeEach(func() {
		err = nil
		appConfig = userconfig.AppDefinition{}

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
		                    { "path": "/data", "size": "5 GB" },
		                    { "path": "/data2", "size": "8" },
		                    { "path": "/data3", "size": "8  G" },
		                    { "path": "/data4", "size": "8GB" }
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

			It("should properly parse given app name", func() {
				Expect(appConfig.AppName).To(Equal("test-app-name"))
			})

			It("should parse one service", func() {
				Expect(appConfig.Services).To(HaveLen(1))
			})

			It("should parse one service with the name 'session'", func() {
				Expect(appConfig.Services[0].ServiceName).To(Equal("session"))
			})

			It("should parse two components", func() {
				Expect(appConfig.Services[0].Components).To(HaveLen(2))
			})

			It("should parse one domain for component", func() {
				Expect(appConfig.Services[0].Components[0].Domains).To(HaveLen(1))
			})

			It("should parse correct component domain", func() {
				Expect(appConfig.Services[0].Components[0].Domains["test.domain.io"].Port).To(Equal("80"))
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

		Describe("parsing app-config with invalid volume size 1", func() {
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
		                    { "path": "/data1", "size": "5KB" }
		                  ]
		                }
		              ]
		            }
		          ]
		        }`)

				err = json.Unmarshal(byteSlice, &appConfig)
			})

			It("should throw error ErrUnknownJSONField", func() {
				Expect(userconfig.IsErrUnknownJsonField(err)).To(BeFalse())
				Expect(err.Error()).To(Equal(`Cannot parse app config. Invalid size '5KB' detected.`))
			})
		})

		Describe("parsing app-config with invalid volume size 2", func() {
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
		                    { "path": "/data2", "size": "-8KB" }
		                  ]
		                }
		              ]
		            }
		          ]
		        }`)

				err = json.Unmarshal(byteSlice, &appConfig)
			})

			It("should throw error ErrUnknownJSONField", func() {
				Expect(userconfig.IsErrUnknownJsonField(err)).To(BeFalse())
				Expect(err.Error()).To(Equal(`Cannot parse app config. Invalid size '-8KB' detected.`))
			})
		})

		Describe("parsing app-config with unknown field", func() {
			BeforeEach(func() {
				byteSlice = []byte(`{
		          "app_name": "test-app-name",
		          "foo":1,
		          "services": [
		            {
		              "service_name": "session"
		            }
		          ]
		        }`)

				err = json.Unmarshal(byteSlice, &appConfig)
			})

			It("should throw error ErrUnknownJSONField", func() {
				Expect(userconfig.IsErrUnknownJsonField(err)).To(BeTrue())
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

				err = json.Unmarshal(byteSlice, &appConfig)
			})

			It("should detect first occuring error and throw ErrUnknownJSONField", func() {
				Expect(userconfig.IsErrUnknownJsonField(err)).To(BeTrue())
				Expect(err.Error()).To(Equal(`Cannot parse app config. Unknown field '["services"][0]["components"][1]["volume"]' detected.`))
			})

			It("should not parse given app name", func() {
				Expect(appConfig.AppName).To(Equal(""))
			})
		})

		Describe("parsing app-config with duplicate volume paths", func() {
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
		                  "volumes": [
		                    { "path": "/data", "size": "5 GB" },
		                    { "path": "/data", "size": "10 GB" }
		                   ]
		                }
		              ]
		            }
		          ]
		        }`)

				err = json.Unmarshal(byteSlice, &appConfig)
			})

			It("should detect first occuring error and throw IsErrDuplicateVolumePath", func() {
				Expect(userconfig.IsErrDuplicateVolumePath(err)).To(BeTrue())
				Expect(err.Error()).To(Equal(`Cannot parse app config. Duplicate volume '/data' found in component 'api'.`))
			})

			It("should not parse given app name", func() {
				Expect(appConfig.AppName).To(Equal(""))
			})
		})

		Context("fix app-config fields", func() {
			Context("ComponentConfig", func() {
				var componentConfig userconfig.ComponentConfig
				BeforeEach(func() {
					componentConfig = userconfig.ComponentConfig{}
				})

				Describe("UnmarshalJSON parses deprecated notation", func() {
					BeforeEach(func() {
						// The "Path" is outdated should now be marshaled as "path"
						byteSlice = []byte(`
							{
								"component_name": "foo-bar",
								"volumes": [
									{"Path":"/mnt","Size":"5 GB"}
								],
								"image": "registry.giantswarm.io/userd"
							}
						`)
						err = json.Unmarshal(byteSlice, &componentConfig)
					})
					It("should not throw an error", func() {
						Expect(err).To(BeNil())
					})
					It("should contain one volume with a path=/mnt", func() {
						Expect(len(componentConfig.Volumes)).To(Equal(1))
						Expect(componentConfig.Volumes[0].Path).To(Equal("/mnt"))
					})
				})

				Describe("MarshalJSON makes all fields lower_case", func() {

					var data []byte

					BeforeEach(func() {
						componentConfig.ComponentName = "Test"
						componentConfig.InstanceConfig = userconfig.InstanceConfig{
							Image: generictypes.MustParseDockerImage("registry.giantswarm.io/giantswarm/foobar"),
							Volumes: []userconfig.VolumeConfig{
								{
									Path: "/mnt",
									Size: "5 GB",
								},
							},
						}

						data, err = json.Marshal(componentConfig)
					})

					It("should not throw an error", func() {
						Expect(err).To(BeNil())
					})

					It("should lowercase the path field", func() {
						Expect(strings.Contains(string(data), "\"path\":")).To(BeTrue())
					})

				})

			})
			Context("AppDefinition", func() {
				Describe("with valid field names", func() {
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
						Expect(appConfig.Services[0].Components[0].Domains["test.domain.io"].Port).To(Equal("80"))
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

				Describe("with deprecated field names", func() {
					BeforeEach(func() {
						// Components was the old name in the ServiceConfig
						byteSlice = []byte(`{
			            "app_name": "test-app-name",
			            "services": [
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
						Expect(appConfig.Services[0].Components[0].Domains["test.domain.io"].Port).To(Equal("80"))
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

				Describe("with invalid field names", func() {
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

						err = json.Unmarshal(byteSlice, &appConfig)
					})

					It("should throw error", func() {
						Expect(userconfig.IsErrUnknownJsonField(err)).To(BeTrue())
						Expect(err.Error()).To(Equal(`Cannot parse app config. Unknown field '["services"][0]["comp_onents"]' detected.`))
					})

					It("should not parse given app name", func() {
						Expect(appConfig.AppName).To(Equal(""))
					})
				})

				Describe("MarshalJSON makes all fields lower_case", func() {

					var data []byte

					BeforeEach(func() {
						appConfig.AppName = "Test"
						appConfig.Services = []userconfig.ServiceConfig{
							{
								ServiceName: "test-service-1",
								Components: []userconfig.ComponentConfig{
									{
										ComponentName: "test-service-1-component-1",
										InstanceConfig: userconfig.InstanceConfig{
											Image: generictypes.MustParseDockerImage("registry.giantswarm.io/giantswarm/foobar"),
										},
									},
								},
							},
						}
						data, err = json.Marshal(appConfig)
					})

					It("should not throw an error", func() {
						Expect(err).To(BeNil())
					})

					It("should lowercase the components field", func() {
						Expect(strings.Contains(string(data), "\"components\":")).To(BeTrue())
					})

				})
			})

		})

	})
})
