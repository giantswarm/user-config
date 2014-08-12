PROJECT=user-config

BUILD_PATH := $(shell pwd)/.gobuild

D0_PATH := "$(BUILD_PATH)/src/github.com/giantswarm"

BIN=$(PROJECT)

.PHONY=clean run-test get-deps update-deps fmt run-tests upload-docker-image

GOPATH := $(BUILD_PATH)

SOURCE=$(shell find . -name '*.go')

all: get-deps $(BIN)

clean:
	rm -rf $(BUILD_PATH) $(BIN)

get-deps: .gobuild

.gobuild:
	mkdir -p $(D0_PATH)
	cd "$(D0_PATH)" && ln -s ../../../.. $(PROJECT)

	#
	# Fetch public dependencies via `go get`
	GOPATH=$(GOPATH) go get -d -v github.com/giantswarm/$(PROJECT)

$(BIN): $(SOURCE)
	GOPATH=$(GOPATH) go build -o $(BIN)

run-tests:
	GOPATH=$(GOPATH) go test ./...

fmt:
	gofmt -l -w .
