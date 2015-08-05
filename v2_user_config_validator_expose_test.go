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
	testComponent := func() *ComponentDefinition {
		return &ComponentDefinition{
			Image: MustParseImageDefinition("registry/namespace/repository:version"),
		}
	}

	addExpose := func(config *ComponentDefinition, exp ...ExposeDefinition) *ComponentDefinition {
		if config.Expose == nil {
			config.Expose = exp
		} else {
			config.Expose = append(config.Expose, exp...)
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

	testApp := func() ComponentDefinitions {
		return make(ComponentDefinitions)
	}

	port := func(p string) generictypes.DockerPort {
		return generictypes.MustParseDockerPort(p)
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

	Describe("intra-app tests", func() {
		Describe("check relations between expose and defined ports", func() {
			Describe("test valid expose implemented by child component with specific implementation port", func() {
				var err error

				BeforeEach(func() {
					components := testApp()
					components["a"] = addExpose(testComponent(), ExposeDefinition{Port: port("123"), Component: "a/b", TargetPort: port("456")})
					components["a/b"] = addPorts(testComponent(), port("456"))

					err = validate(components)
				})

				It("should not throw an error", func() {
					Expect(err).To(BeNil())
				})
			})

			Describe("test valid expose implemented by child component without implementation port", func() {
				var err error

				BeforeEach(func() {
					components := testApp()
					components["a"] = addExpose(testComponent(), ExposeDefinition{Port: port("123"), Component: "a/b/c"})
					components["a/b/c"] = addPorts(testComponent(), port("123"))

					err = validate(components)
				})

				It("should not throw an error", func() {
					Expect(err).To(BeNil())
				})
			})

			Describe("test valid expose implemented by same component with specific implementation port", func() {
				var err error

				BeforeEach(func() {
					components := testApp()
					components["a"] = addPorts(addExpose(testComponent(), ExposeDefinition{Port: port("123"), TargetPort: port("456")}), port("456"))

					err = validate(components)
				})

				It("should not throw an error", func() {
					Expect(err).To(BeNil())
				})
			})

			Describe("test invalid expose implemented by non-child component with specific implementation port", func() {
				var err error

				BeforeEach(func() {
					components := testApp()
					components["a"] = addExpose(testComponent(), ExposeDefinition{Port: port("123"), Component: "b", TargetPort: port("456")})
					components["b"] = addPorts(testComponent(), port("456"))

					err = validate(components)
				})

				It("should throw an InvalidComponentDefinitionError", func() {
					Expect(IsInvalidComponentDefinition(err)).To(BeTrue())
					Expect(err.Error()).To(Equal(`invalid expose to component 'b': is not a child of 'a'`))
				})
			})

			Describe("test invalid expose implemented by child component with specific implementation port that does not exist in port list", func() {
				var err error

				BeforeEach(func() {
					components := testApp()
					components["a"] = addExpose(testComponent(), ExposeDefinition{Port: port("123"), Component: "a/b", TargetPort: port("456")})
					components["a/b"] = addPorts(testComponent(), port("789"))

					err = validate(components)
				})

				It("should throw an InvalidComponentDefinitionError", func() {
					Expect(IsInvalidComponentDefinition(err)).To(BeTrue())
					Expect(err.Error()).To(Equal(`invalid expose to component 'a/b': does not export port '456/tcp'`))
				})
			})

			Describe("test invalid expose implemented by child component without specific implementation port that does not exist in port list", func() {
				var err error

				BeforeEach(func() {
					components := testApp()
					components["a"] = addExpose(testComponent(), ExposeDefinition{Port: port("123"), Component: "a/b"})
					components["a/b"] = testComponent()

					err = validate(components)
				})

				It("should throw an InvalidComponentDefinitionError", func() {
					Expect(IsInvalidComponentDefinition(err)).To(BeTrue())
					Expect(err.Error()).To(Equal(`invalid expose to component 'a/b': does not export port '123/tcp'`))
				})
			})

			Describe("test invalid duplicate expose of the same port", func() {
				var err error

				BeforeEach(func() {
					components := testApp()
					components["a"] = addExpose(testComponent(),
						ExposeDefinition{Port: port("123"), Component: "a/b"},
						ExposeDefinition{Port: port("123"), Component: "a/b"},
					)
					components["a/b"] = testComponent()

					err = validate(components)
				})

				It("should throw an InvalidComponentDefinitionError", func() {
					Expect(IsInvalidComponentDefinition(err)).To(BeTrue())
					Expect(err.Error()).To(Equal(`port '123/tcp' is exposed more than once`))
				})
			})
		})

		Describe("check relations between expose and links to that expose", func() {
			Describe("test valid link to an exposed api", func() {
				var err error

				BeforeEach(func() {
					components := testApp()
					components["a"] = addExpose(testComponent(), ExposeDefinition{Port: port("123"), Component: "a/b", TargetPort: port("456")})
					components["a/b"] = addPorts(testComponent(), port("456"))
					components["c"] = addLinks(testComponent(), LinkDefinition{Component: "a", TargetPort: port("123")})

					err = validate(components)
				})

				It("should not throw an error", func() {
					Expect(err).To(BeNil())
				})
			})

			Describe("test invalid link to a non-exposed api", func() {
				var err error

				BeforeEach(func() {
					components := testApp()
					components["a"] = addExpose(testComponent(), ExposeDefinition{Port: port("123"), Component: "a/b", TargetPort: port("456")})
					components["a/b"] = addPorts(testComponent(), port("456"))
					components["c"] = addLinks(testComponent(), LinkDefinition{Component: "a", TargetPort: port("789")})

					err = validate(components)
				})

				It("should throw an InvalidComponentDefinitionError", func() {
					Expect(IsInvalidComponentDefinition(err)).To(BeTrue())
					Expect(err.Error()).To(Equal(`invalid link to component 'a': does not export port '789/tcp'`))
				})
			})
		})

		Describe("check restrictions for inter-service links", func() {
			Describe("test valid link to another service", func() {
				var err error

				BeforeEach(func() {
					components := testApp()
					components["a"] = addLinks(testComponent(), LinkDefinition{Service: "other", TargetPort: port("123")})

					err = validate(components)
				})

				It("should not throw an error", func() {
					Expect(err).To(BeNil())
				})
			})

			Describe("test invalid link; cannot have name and service set", func() {
				var err error

				BeforeEach(func() {
					components := testApp()
					components["a"] = addLinks(testComponent(), LinkDefinition{Service: "other", Component: "b", TargetPort: port("123")})
					components["b"] = addPorts(testComponent(), port("123"))

					err = validate(components)
				})

				It("should throw an InvalidLinkDefinitionError", func() {
					Expect(IsInvalidLinkDefinition(err)).To(BeTrue())
					Expect(err.Error()).To(Equal(`link service and component cannot be set both`))
				})
			})

			Describe("test exposing the same port on multiple root components (which is not allowed)", func() {
				var err error

				BeforeEach(func() {
					components := testApp()
					components["a"] = addExpose(testComponent(), ExposeDefinition{Port: port("123"), Component: "a/b", TargetPort: port("456")})
					components["a/b"] = addPorts(testComponent(), port("456"))
					components["c"] = addExpose(testComponent(), ExposeDefinition{Port: port("123"), Component: "c/b", TargetPort: port("456")})
					components["c/b"] = addPorts(testComponent(), port("456"))

					err = validate(components)
				})

				It("should throw an InvalidComponentDefinitionError", func() {
					Expect(IsInvalidComponentDefinition(err)).To(BeTrue())
					Expect(err.Error()).To(Equal(`port '123/tcp' is exposed by multiple root components`))
				})
			})
		})

	})
})
