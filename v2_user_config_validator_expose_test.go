package userconfig

import (
	"testing"

	"github.com/giantswarm/generic-types-go"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestV2UserConfigExposeValidator(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "v2 user config stable API validator")
}

var _ = Describe("v2 user config stable API validator", func() {

	testNode := func() *NodeDefinition {
		return &NodeDefinition{
			Image: MustParseImageDefinition("registry/namespace/repository:version"),
		}
	}

	addExpose := func(config *NodeDefinition, exp ...ExposeDefinition) *NodeDefinition {
		if config.Expose == nil {
			config.Expose = exp
		} else {
			config.Expose = append(config.Expose, exp...)
		}
		return config
	}

	addLinks := func(config *NodeDefinition, link ...LinkDefinition) *NodeDefinition {
		if config.Links == nil {
			config.Links = link
		} else {
			config.Links = append(config.Links, link...)
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

	testApp := func() NodeDefinitions {
		return make(NodeDefinitions)
	}

	port := func(p string) generictypes.DockerPort {
		return generictypes.MustParseDockerPort(p)
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

	Describe("intra-app tests", func() {
		Describe("check relations between expose and defined ports", func() {
			Describe("test valid expose implemented by child node with specific implementation port", func() {
				var err error

				BeforeEach(func() {
					nodes := testApp()
					nodes["a"] = addExpose(testNode(), ExposeDefinition{Port: port("123"), Node: "a/b", NodePort: port("456")})
					nodes["a/b"] = addPorts(testNode(), port("456"))

					err = validate(nodes)
				})

				It("should not throw an error", func() {
					Expect(err).To(BeNil())
				})
			})

			Describe("test valid expose implemented by child node without implementation port", func() {
				var err error

				BeforeEach(func() {
					nodes := testApp()
					nodes["a"] = addExpose(testNode(), ExposeDefinition{Port: port("123"), Node: "a/b/c"})
					nodes["a/b/c"] = addPorts(testNode(), port("123"))

					err = validate(nodes)
				})

				It("should not throw an error", func() {
					Expect(err).To(BeNil())
				})
			})

			Describe("test valid expose implemented by same node with specific implementation port", func() {
				var err error

				BeforeEach(func() {
					nodes := testApp()
					nodes["a"] = addPorts(addExpose(testNode(), ExposeDefinition{Port: port("123"), NodePort: port("456")}), port("456"))

					err = validate(nodes)
				})

				It("should not throw an error", func() {
					Expect(err).To(BeNil())
				})
			})

			Describe("test invalid expose implemented by non-child node with specific implementation port", func() {
				var err error

				BeforeEach(func() {
					nodes := testApp()
					nodes["a"] = addExpose(testNode(), ExposeDefinition{Port: port("123"), Node: "b", NodePort: port("456")})
					nodes["b"] = addPorts(testNode(), port("456"))

					err = validate(nodes)
				})

				It("should throw an InvalidNodeDefinitionError", func() {
					Expect(IsInvalidNodeDefinition(err)).To(BeTrue())
					Expect(err.Error()).To(Equal(`invalid expose to node 'b': is not a child of 'a'`))
				})
			})

			Describe("test invalid expose implemented by child node with specific implementation port that does not exist in port list", func() {
				var err error

				BeforeEach(func() {
					nodes := testApp()
					nodes["a"] = addExpose(testNode(), ExposeDefinition{Port: port("123"), Node: "a/b", NodePort: port("456")})
					nodes["a/b"] = addPorts(testNode(), port("789"))

					err = validate(nodes)
				})

				It("should throw an InvalidNodeDefinitionError", func() {
					Expect(IsInvalidNodeDefinition(err)).To(BeTrue())
					Expect(err.Error()).To(Equal(`invalid expose to node 'a/b': does not export port '456/tcp'`))
				})
			})

			Describe("test invalid expose implemented by child node without specific implementation port that does not exist in port list", func() {
				var err error

				BeforeEach(func() {
					nodes := testApp()
					nodes["a"] = addExpose(testNode(), ExposeDefinition{Port: port("123"), Node: "a/b"})
					nodes["a/b"] = testNode()

					err = validate(nodes)
				})

				It("should throw an InvalidNodeDefinitionError", func() {
					Expect(IsInvalidNodeDefinition(err)).To(BeTrue())
					Expect(err.Error()).To(Equal(`invalid expose to node 'a/b': does not export port '123/tcp'`))
				})
			})

			Describe("test invalid duplicate expose of the same port", func() {
				var err error

				BeforeEach(func() {
					nodes := testApp()
					nodes["a"] = addExpose(testNode(),
						ExposeDefinition{Port: port("123"), Node: "a/b"},
						ExposeDefinition{Port: port("123"), Node: "a/b"},
					)
					nodes["a/b"] = testNode()

					err = validate(nodes)
				})

				It("should throw an InvalidNodeDefinitionError", func() {
					Expect(IsInvalidNodeDefinition(err)).To(BeTrue())
					Expect(err.Error()).To(Equal(`port '123/tcp' is exposed more than once`))
				})
			})
		})

		Describe("check relations between expose and links to that expose", func() {
			Describe("test valid link to an exposed api", func() {
				var err error

				BeforeEach(func() {
					nodes := testApp()
					nodes["a"] = addExpose(testNode(), ExposeDefinition{Port: port("123"), Node: "a/b", NodePort: port("456")})
					nodes["a/b"] = addPorts(testNode(), port("456"))
					nodes["c"] = addLinks(testNode(), LinkDefinition{Name: "a", Port: port("123")})

					err = validate(nodes)
				})

				It("should not throw an error", func() {
					Expect(err).To(BeNil())
				})
			})

			Describe("test invalid link to a non-exposed api", func() {
				var err error

				BeforeEach(func() {
					nodes := testApp()
					nodes["a"] = addExpose(testNode(), ExposeDefinition{Port: port("123"), Node: "a/b", NodePort: port("456")})
					nodes["a/b"] = addPorts(testNode(), port("456"))
					nodes["c"] = addLinks(testNode(), LinkDefinition{Name: "a", Port: port("789")})

					err = validate(nodes)
				})

				It("should throw an InvalidNodeDefinitionError", func() {
					Expect(IsInvalidNodeDefinition(err)).To(BeTrue())
					Expect(err.Error()).To(Equal(`invalid link to node 'a': does not export port '789/tcp'`))
				})
			})
		})

		Describe("check restrictions for inter-app links", func() {
			Describe("test valid link to another app", func() {
				var err error

				BeforeEach(func() {
					nodes := testApp()
					nodes["a"] = addLinks(testNode(), LinkDefinition{App: "other", Port: port("123")})

					err = validate(nodes)
				})

				It("should not throw an error", func() {
					Expect(err).To(BeNil())
				})
			})

			Describe("test invalid link; cannot have name and app set", func() {
				var err error

				BeforeEach(func() {
					nodes := testApp()
					nodes["a"] = addLinks(testNode(), LinkDefinition{App: "other", Name: "b", Port: port("123")})
					nodes["b"] = addPorts(testNode(), port("123"))

					err = validate(nodes)
				})

				It("should throw an InvalidLinkDefinitionError", func() {
					Expect(IsInvalidLinkDefinition(err)).To(BeTrue())
					Expect(err.Error()).To(Equal(`link app and name cannot be set both`))
				})
			})

			Describe("test exposing the same port on multiple root nodes (which is not allowed)", func() {
				var err error

				BeforeEach(func() {
					nodes := testApp()
					nodes["a"] = addExpose(testNode(), ExposeDefinition{Port: port("123"), Node: "a/b", NodePort: port("456")})
					nodes["a/b"] = addPorts(testNode(), port("456"))
					nodes["c"] = addExpose(testNode(), ExposeDefinition{Port: port("123"), Node: "c/b", NodePort: port("456")})
					nodes["c/b"] = addPorts(testNode(), port("456"))

					err = validate(nodes)
				})

				It("should throw an InvalidNodeDefinitionError", func() {
					Expect(IsInvalidNodeDefinition(err)).To(BeTrue())
					Expect(err.Error()).To(Equal(`port '123/tcp' is exposed by multiple root nodes`))
				})
			})
		})

	})
})
