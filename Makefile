PROJECT=user-config

BUILD_PATH := $(shell pwd)/.gobuild

GS_PATH := "$(BUILD_PATH)/src/github.com/giantswarm"

BIN=$(PROJECT)

.PHONY=clean run-test get-deps update-deps fmt run-tests upload-docker-image

GOPATH := $(BUILD_PATH)

SOURCE=$(shell find . -name '*.go')

all: get-deps $(BIN)

ci: clean all run-tests

clean:
	rm -rf $(BUILD_PATH) $(BIN)

get-deps: .gobuild

.gobuild:
	mkdir -p $(GS_PATH)
	cd "$(GS_PATH)" && ln -s ../../../.. $(PROJECT)

	#
	# Fetch internal libraries
	@builder get dep -b empty git@github.com:giantswarm/generic-types-go.git $(GS_PATH)/generic-types-go
	@builder get dep git@github.com:giantswarm/go-tld.git $(GS_PATH)/go-tld
	@builder get dep git@github.com:giantswarm/validate.git $(GS_PATH)/validate

	#
	# Fetch public dependencies via `go get`
	GOPATH=$(GOPATH) builder go get github.com/juju/errgo
	GOPATH=$(GOPATH) builder go get github.com/kr/pretty
	GOPATH=$(GOPATH) builder go get github.com/kr/text

	#
	# Build test packages (we only want those two, so we use `-d` in go get)
	GOPATH=$(GOPATH) builder go get github.com/onsi/gomega
	GOPATH=$(GOPATH) builder go get github.com/onsi/ginkgo

$(BIN): $(SOURCE)
	GOPATH=$(GOPATH) go build -o $(BIN)

run-tests:
	GOPATH=$(GOPATH) go test ./...

run-test:
	if test "$(test)" = "" ; then \
		echo "missing test parameter, that is, path to test folder e.g. './middleware/v1/'."; \
		exit 1; \
	fi
	GOPATH=$(GOPATH) go test -v $(test)

fmt:
	gofmt -l -w .
