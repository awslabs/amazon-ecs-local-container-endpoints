# Copyright 2019 Amazon.com, Inc. or its affiliates. All Rights Reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License"). You
# may not use this file except in compliance with the License. A copy of
# the License is located at
#
# 	http://aws.amazon.com/apache2.0/
#
# or in the "license" file accompanying this file. This file is
# distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF
# ANY KIND, either express or implied. See the License for the specific
# language governing permissions and limitations under the License.

ROOT := $(shell pwd)

all: local-build

GO_VERSION := 1.15
SCRIPT_PATH := $(ROOT)/scripts/:${PATH}
SOURCES := $(shell find . -name '*.go')
BINARY_NAME := local-container-endpoints
IMAGE_NAME := amazon/amazon-ecs-local-container-endpoints
LOCAL_BINARY := bin/local/${BINARY_NAME}
AMD_DIR := linux-amd64
ARM_DIR := linux-arm64
AMD_BINARY := bin/${AMD_DIR}/${BINARY_NAME}
ARM_BINARY := bin/${ARM_DIR}/${BINARY_NAME}
VERSION := $(shell cat VERSION)
AGENT_VERSION_COMPATIBILITY := $(shell cat AGENT_VERSION_COMPATIBILITY)
TAG := $(VERSION)-agent$(AGENT_VERSION_COMPATIBILITY)-compatible

.PHONY: local-build
local-build: $(LOCAL_BINARY)

# build binaries for each architecture into their own subdirectories
.PHONY: linux-compile
linux-compile: $(AMD_BINARY) $(ARM_BINARY)

$(LOCAL_BINARY): $(SOURCES)
	PATH=${PATH} golint ./local-container-endpoints/...
	./scripts/build_binary.sh ./bin/local
	@echo "Built local-container-endpoints"

.PHONY: generate
generate: $(SOURCES)
	PATH=$(SCRIPT_PATH) go generate ./...

.PHONY: test
test:
	go test -mod=vendor -timeout=120s -v -cover ./local-container-endpoints/...

.PHONY: functional-test
functional-test:
	go test -mod=vendor -timeout=120s -v -tags functional -cover ./local-container-endpoints/handlers/functional_tests/...

$(AMD_BINARY): $(SOURCES)
	@mkdir -p ./bin/$(AMD_DIR)
	TARGET_GOOS=linux GOARCH=amd64 ./scripts/build_binary.sh ./bin/$(AMD_DIR)
	@echo "Built local-container-endpoints for linux-amd64"

$(ARM_BINARY): $(SOURCES)
	@mkdir -p ./bin/$(ARM_DIR)
	TARGET_GOOS=linux GOARCH=arm64 ./scripts/build_binary.sh ./bin/$(ARM_DIR)
	@echo "Built local-container-endpoints for linux-arm64"

# release uses each architecture-specific go binary to build images
.PHONY: release
release: release-amd release-arm

.PHONY: release-amd
release-amd:
	docker run -v $(shell pwd):/usr/src/app/src/github.com/awslabs/amazon-ecs-local-container-endpoints \
		--workdir=/usr/src/app/src/github.com/awslabs/amazon-ecs-local-container-endpoints \
		--env GOPATH=/usr/src/app \
		--env ECS_RELEASE=cleanbuild \
		golang:$(GO_VERSION) make $(AMD_BINARY)
	docker build --build-arg ARCH_DIR=$(AMD_DIR) -t $(IMAGE_NAME):latest-amd64 .
	docker tag $(IMAGE_NAME):latest-amd64 $(IMAGE_NAME):$(TAG)-amd64
	docker tag $(IMAGE_NAME):latest-amd64 $(IMAGE_NAME):$(VERSION)-amd64

.PHONY: release-arm
release-arm:
	docker run -v $(shell pwd):/usr/src/app/src/github.com/awslabs/amazon-ecs-local-container-endpoints \
		--workdir=/usr/src/app/src/github.com/awslabs/amazon-ecs-local-container-endpoints \
		--env GOPATH=/usr/src/app \
		--env ECS_RELEASE=cleanbuild \
		golang:$(GO_VERSION) make $(ARM_BINARY)
	docker build --build-arg ARCH_DIR=$(ARM_DIR) -t $(IMAGE_NAME):latest-arm64 .
	docker tag $(IMAGE_NAME):latest-arm64 $(IMAGE_NAME):$(TAG)-arm64
	docker tag $(IMAGE_NAME):latest-arm64 $(IMAGE_NAME):$(VERSION)-arm64

.PHONY: integ
integ: release
	docker build -t amazon-ecs-local-container-endpoints-integ-test:latest -f ./integ/Dockerfile .
	docker-compose --file ./integ/docker-compose.yml up --abort-on-container-exit

.PHONY: publish
publish: release publish-amd publish-arm

.PHONY: publish-amd
publish-amd:
	docker push $(IMAGE_NAME):latest-amd64
	docker push $(IMAGE_NAME):$(TAG)-amd64
	docker push $(IMAGE_NAME):$(VERSION)-amd64

.PHONY: publish-arm
publish-arm:
	docker push $(IMAGE_NAME):latest-arm64
	docker push $(IMAGE_NAME):$(TAG)-arm64
	docker push $(IMAGE_NAME):$(VERSION)-arm64

.PHONY: clean
clean:
	rm bin/local/local-container-endpoints
