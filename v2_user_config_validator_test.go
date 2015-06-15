package userconfig_test

import (
	"encoding/json"
	"reflect"
	"testing"

	//"github.com/giantswarm/generic-types-go"
	"github.com/giantswarm/user-config"
)

func TestMarshalUnmarshalV2AppDef(t *testing.T) {
	a := V2ExampleDefinition()

	data, err := json.Marshal(a)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}

	var b userconfig.V2AppDefinition
	if err := json.Unmarshal(data, &b); err != nil {
		t.Fatalf("json.Unmarshal failed: %v", err)
	}

	if !reflect.DeepEqual(a, b) {
		t.Fatalf("objects differ:\n%v\n%v", a, b)
	}
}

func TestParseV2AppDef(t *testing.T) {
	b := []byte(`{
		"nodes": {
			"node/a": {
				"image": "registry/namespace/repository:version",
				"ports": [ "80/tcp" ],
				"links": [
					{ "name": "redis", "port": 6379, "same_machine": true }
				],
				"domains": { "test.domain.io": "80" }
			},
			"node/b": {
				"image": "dockerfile/redis",
				"ports": [ "6379/tcp" ],
				"volumes": [
					{ "path": "/data", "size": "5 GB" },
					{ "path": "/data2", "size": "8" },
					{ "path": "/data3", "size": "8  G" },
					{ "path": "/data4", "size": "8GB" }
				]
			}
		}
	}`)

	var appDef userconfig.V2AppDefinition
	err := json.Unmarshal(b, &appDef)
	if err != nil {
		t.Fatalf("json.Unmarshal failed: %v", err)
	}

	if len(appDef.Nodes) != 2 {
		t.Fatalf("expected two nodes: %d given", len(appDef.Nodes))
	}

	nodeA, ok := appDef.Nodes["node/a"]
	if !ok {
		t.Fatalf("missing node")
	}

	if len(nodeA.Domains) != 1 {
		t.Fatalf("expected one domain: %d given", len(nodeA.Domains))
	}

	port, ok := nodeA.Domains["test.domain.io"]
	if !ok {
		t.Fatalf("missing domain")
	}
	if port.String() != "80/tcp" {
		t.Fatalf("invalid port: %s", port.String())
	}

	if nodeA.Image.Registry != "registry" {
		t.Fatalf("invalid registry: %s", nodeA.Image.Registry)
	}
	if nodeA.Image.Namespace != "namespace" {
		t.Fatalf("invalid namespace: %s", nodeA.Image.Namespace)
	}
	if nodeA.Image.Repository != "repository" {
		t.Fatalf("invalid repository: %s", nodeA.Image.Repository)
	}
	if nodeA.Image.Version != "version" {
		t.Fatalf("invalid version: %s", nodeA.Image.Version)
	}

	nodeB, ok := appDef.Nodes["node/b"]
	if !ok {
		t.Fatalf("missing node")
	}

	if nodeB.Image.Registry != "" {
		t.Fatalf("invalid registry: %s", nodeB.Image.Registry)
	}
	if nodeB.Image.Namespace != "dockerfile" {
		t.Fatalf("invalid namespace: %s", nodeB.Image.Namespace)
	}
	if nodeB.Image.Repository != "redis" {
		t.Fatalf("invalid repository: %s", nodeB.Image.Repository)
	}
	if nodeB.Image.Version != "" {
		t.Fatalf("invalid version: %s", nodeB.Image.Version)
	}
}

func TestV2AppDefInvalidVolumeSizeUnit(t *testing.T) {
	a := V2ExampleDefinitionWithVolume("/data", "5 KB")

	raw, err := json.Marshal(a)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}

	var b userconfig.V2AppDefinition
	err = json.Unmarshal(raw, &b)
	if err == nil {
		t.Fatalf("json.Unmarshal NOT failed")
	}

	if err.Error() != "Cannot parse app config. Invalid size '5 KB' detected." {
		t.Fatalf("expected proper error, got: %s", err.Error())
	}
}

func TestV2AppDefInvalidVolumeNegativeSize(t *testing.T) {
	a := V2ExampleDefinitionWithVolume("/data", "-5 GB")

	raw, err := json.Marshal(a)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}

	var b userconfig.V2AppDefinition
	err = json.Unmarshal(raw, &b)
	if err == nil {
		t.Fatalf("json.Unmarshal NOT failed")
	}

	if err.Error() != "Cannot parse app config. Invalid size '-5 GB' detected." {
		t.Fatalf("expected proper error, got: %s", err.Error())
	}
}

func TestV2AppDefInvalidFieldName(t *testing.T) {
	b := []byte(`{
		"nodes": {
			"node/a": {
				"image": "registry/namespace/repository:version",
				"foo": [ "80/tcp" ]
			}
		}
	}`)

	var appDef userconfig.V2AppDefinition
	err := json.Unmarshal(b, &appDef)
	if err == nil {
		t.Fatalf("json.Unmarshal NOT failed")
	}
	if err.Error() != `Cannot parse app definition. Unknown field '["nodes"]["node/a"]["foo"]' detected.` {
		t.Fatalf("expected proper error, got: %s", err.Error())
	}
	if !userconfig.IsErrUnknownJsonField(err) {
		t.Fatalf("expetced error to be ErrUnknownJSONField")
	}
}

//	Describe("parsing app-config with duplicate volume paths", func() {
//		BeforeEach(func() {
//			byteSlice = []byte(`{
//	          "app_name": "test-app-name",
//	          "services": [
//	            {
//	              "service_name": "session",
//	              "components": [
//	                {
//	                  "component_name": "api",
//	                  "image": "registry/namespace/repository:version",
//	                  "volumes": [
//	                    { "path": "/data", "size": "5 GB" },
//	                    { "path": "/data", "size": "10 GB" }
//	                   ]
//	                }
//	              ]
//	            }
//	          ]
//	        }`)

//			err = json.Unmarshal(byteSlice, &appConfig)
//		})

//		It("should detect first occuring error and throw IsErrDuplicateVolumePath", func() {
//			Expect(userconfig.IsErrDuplicateVolumePath(err)).To(BeTrue())
//			Expect(err.Error()).To(Equal(`Cannot parse app config. Duplicate volume '/data' detected.`))
//		})

//		It("should not parse given app name", func() {
//			Expect(appConfig.AppName).To(Equal(""))
//		})
//	})

//	Context("fix app-config fields", func() {
//		Context("ComponentConfig", func() {
//			var componentConfig userconfig.ComponentConfig
//			BeforeEach(func() {
//				componentConfig = userconfig.ComponentConfig{}
//			})

//			Describe("UnmarshalJSON parses deprecated notation", func() {
//				BeforeEach(func() {
//					// The "Path" is outdated should now be marshaled as "path"
//					byteSlice = []byte(`
//						{
//							"component_name": "foo-bar",
//							"volumes": [
//								{"Path":"/mnt","Size":"5 GB"}
//							],
//							"image": "registry.giantswarm.io/userd"
//						}
//					`)
//					err = json.Unmarshal(byteSlice, &componentConfig)
//				})
//				It("should not throw an error", func() {
//					Expect(err).To(BeNil())
//				})
//				It("should contain one volume with a path=/mnt", func() {
//					Expect(len(componentConfig.Volumes)).To(Equal(1))
//					Expect(componentConfig.Volumes[0].Path).To(Equal("/mnt"))
//				})
//			})

//			Describe("MarshalJSON makes all fields lower_case", func() {

//				var data []byte

//				BeforeEach(func() {
//					componentConfig.ComponentName = "Test"
//					componentConfig.InstanceConfig = userconfig.InstanceConfig{
//						Image: generictypes.MustParseDockerImage("registry.giantswarm.io/giantswarm/foobar"),
//						Volumes: []userconfig.VolumeConfig{
//							{
//								Path: "/mnt",
//								Size: "5 GB",
//							},
//						},
//					}

//					data, err = json.Marshal(componentConfig)
//				})

//				It("should not throw an error", func() {
//					Expect(err).To(BeNil())
//				})

//				It("should lowercase the path field", func() {
//					Expect(strings.Contains(string(data), "\"path\":")).To(BeTrue())
//				})

//			})

//		})
//		Context("AppDefinition", func() {
//			Describe("with valid field names", func() {
//				BeforeEach(func() {
//					byteSlice = []byte(`{
//		            "app_name": "test-app-name",
//		            "services": [
//		              {
//		                "service_name": "session",
//		                "components": [
//		                  {
//		                    "component_name": "api",
//		                    "image": "registry/namespace/repository:version",
//		                    "ports": [ "80/tcp" ],
//		                    "dependencies": [
//		                      { "name": "redis", "port": 6379, "same_machine": true }
//		                    ],
//		                    "domains": { "test.domain.io": "80" }
//		                  },
//		                  {
//		                    "component_name": "redis",
//		                    "image": "dockerfile/redis",
//		                    "ports": [ "6379/tcp" ],
//		                    "volumes": [
//		                      { "path": "/data", "size": "5 GB" }
//		                    ]
//		                  }
//		                ]
//		              }
//		            ]
//		          }`)

//					err = json.Unmarshal(byteSlice, &appConfig)

//				})

//				It("should not throw error", func() {
//					Expect(err).To(BeNil())
//				})

//				It("should properly parse given app name", func() {
//					Expect(appConfig.AppName).To(Equal("test-app-name"))
//				})

//				It("should parse one service", func() {
//					Expect(appConfig.Services).To(HaveLen(1))
//				})

//				It("should parse one service", func() {
//					Expect(appConfig.Services[0].ServiceName).To(Equal("session"))
//				})

//				It("should parse two components", func() {
//					Expect(appConfig.Services[0].Components).To(HaveLen(2))
//				})

//				It("should parse one domain for component", func() {
//					Expect(appConfig.Services[0].Components[0].Domains).To(HaveLen(1))
//				})

//				It("should parse correct component domain", func() {
//					Expect(appConfig.Services[0].Components[0].Domains["test.domain.io"].Port).To(Equal("80"))
//				})

//				It("should parse correct component image 1", func() {
//					Expect(appConfig.Services[0].Components[0].Image.Registry).To(Equal("registry"))
//					Expect(appConfig.Services[0].Components[0].Image.Namespace).To(Equal("namespace"))
//					Expect(appConfig.Services[0].Components[0].Image.Repository).To(Equal("repository"))
//					Expect(appConfig.Services[0].Components[0].Image.Version).To(Equal("version"))
//				})

//				It("should parse correct component image 2", func() {
//					Expect(appConfig.Services[0].Components[1].Image.Registry).To(Equal(""))
//					Expect(appConfig.Services[0].Components[1].Image.Namespace).To(Equal("dockerfile"))
//					Expect(appConfig.Services[0].Components[1].Image.Repository).To(Equal("redis"))
//					Expect(appConfig.Services[0].Components[1].Image.Version).To(Equal(""))
//				})
//			})

//			Describe("with deprecated field names", func() {
//				BeforeEach(func() {
//					// Components was the old name in the ServiceConfig
//					byteSlice = []byte(`{
//		            "app_name": "test-app-name",
//		            "services": [
//		              {
//		                "service_name": "session",
//		                "Components": [
//		                  {
//		                    "component_name": "api",
//		                    "image": "registry/namespace/repository:version",
//		                    "ports": [ "80/tcp" ],
//		                    "dependencies": [
//		                      { "name": "redis", "port": 6379, "same_machine": true }
//		                    ],
//		                    "domains": { "test.domain.io": "80" }
//		                  },
//		                  {
//		                    "component_name": "redis",
//		                    "image": "dockerfile/redis",
//		                    "ports": [ "6379/tcp" ],
//		                    "volumes": [
//		                      { "path": "/data", "size": "5 GB" }
//		                    ]
//		                  }
//		                ]
//		              }
//		            ]
//		          }`)

//					err = json.Unmarshal(byteSlice, &appConfig)
//				})

//				It("should not throw error", func() {
//					Expect(err).To(BeNil())
//				})

//				It("should properly parse given app name", func() {
//					Expect(appConfig.AppName).To(Equal("test-app-name"))
//				})

//				It("should parse one service", func() {
//					Expect(appConfig.Services).To(HaveLen(1))
//				})

//				It("should parse one service", func() {
//					Expect(appConfig.Services[0].ServiceName).To(Equal("session"))
//				})

//				It("should parse two components", func() {
//					Expect(appConfig.Services[0].Components).To(HaveLen(2))
//				})

//				It("should parse one domain for component", func() {
//					Expect(appConfig.Services[0].Components[0].Domains).To(HaveLen(1))
//				})

//				It("should parse correct component domain", func() {
//					Expect(appConfig.Services[0].Components[0].Domains["test.domain.io"].Port).To(Equal("80"))
//				})

//				It("should parse correct component image 1", func() {
//					Expect(appConfig.Services[0].Components[0].Image.Registry).To(Equal("registry"))
//					Expect(appConfig.Services[0].Components[0].Image.Namespace).To(Equal("namespace"))
//					Expect(appConfig.Services[0].Components[0].Image.Repository).To(Equal("repository"))
//					Expect(appConfig.Services[0].Components[0].Image.Version).To(Equal("version"))
//				})

//				It("should parse correct component image 2", func() {
//					Expect(appConfig.Services[0].Components[1].Image.Registry).To(Equal(""))
//					Expect(appConfig.Services[0].Components[1].Image.Namespace).To(Equal("dockerfile"))
//					Expect(appConfig.Services[0].Components[1].Image.Repository).To(Equal("redis"))
//					Expect(appConfig.Services[0].Components[1].Image.Version).To(Equal(""))
//				})
//			})

//			Describe("with invalid field names", func() {
//				BeforeEach(func() {
//					byteSlice = []byte(`{
//		            "app_name": "test-app-name",
//		            "services": [
//		              {
//		                "service_name": "session",
//		                "compOnents": [
//		                  {
//		                    "component_name": "api",
//		                    "image": "registry/namespace/repository:version",
//		                    "ports": [ "80/tcp" ],
//		                    "dependencies": [
//		                      { "name": "redis", "port": 6379, "same_machine": true }
//		                    ],
//		                    "domains": { "test.domain.io": "80" }
//		                  },
//		                  {
//		                    "component_name": "redis",
//		                    "image": "dockerfile/redis",
//		                    "ports": [ "6379/tcp" ],
//		                    "volumes": [
//		                      { "path": "/data", "size": "5 GB" }
//		                    ]
//		                  }
//		                ]
//		              }
//		            ]
//		          }`)

//					err = json.Unmarshal(byteSlice, &appConfig)
//				})

//				It("should throw error", func() {
//					Expect(userconfig.IsErrUnknownJsonField(err)).To(BeTrue())
//					Expect(err.Error()).To(Equal(`Cannot parse app config. Unknown field '["services"][0]["comp_onents"]' detected.`))
//				})

//				It("should not parse given app name", func() {
//					Expect(appConfig.AppName).To(Equal(""))
//				})
//			})

//			Describe("MarshalJSON makes all fields lower_case", func() {

//				var data []byte

//				BeforeEach(func() {
//					appConfig.AppName = "Test"
//					appConfig.Services = []userconfig.ServiceConfig{
//						{
//							ServiceName: "test-service-1",
//							Components: []userconfig.ComponentConfig{
//								{
//									ComponentName: "test-service-1-component-1",
//									InstanceConfig: userconfig.InstanceConfig{
//										Image: generictypes.MustParseDockerImage("registry.giantswarm.io/giantswarm/foobar"),
//									},
//								},
//							},
//						},
//					}
//					data, err = json.Marshal(appConfig)
//				})

//				It("should not throw an error", func() {
//					Expect(err).To(BeNil())
//				})

//				It("should lowercase the components field", func() {
//					Expect(strings.Contains(string(data), "\"components\":")).To(BeTrue())
//				})

//			})
//		})

//	})

//})
