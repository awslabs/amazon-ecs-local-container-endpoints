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

GO_VERSION = 1.15
SCRIPT_PATH := $(ROOT)/scripts/:${PATH}
SOURCES := $(shell find . -name '*.go')
BINARY_NAME=local-container-endpoints
LOCAL_BINARY := bin/${BINARY_NAME}
AMD_BINARY := bin/linux-amd64/${BINARY_NAME}
ARM_BINARY := bin/linux-arm64/${BINARY_NAME}
VERSION := $(shell cat VERSION)
AGENT_VERSION_COMPATIBILITY := $(shell cat AGENT_VERSION_COMPATIBILITY)
TAG := $(VERSION)-agent$(AGENT_VERSION_COMPATIBILITY)-compatible

.PHONY: local-build
local-build: $(LOCAL_BINARY)

.PHONY: linux-build
linux-build: $(AMD_BINARY) $(ARM_BINARY)

$(LOCAL_BINARY): $(SOURCES)
	PATH=${PATH} golint ./local-container-endpoints/...
	./scripts/build_binary.sh ./bin/
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
	@mkdir -p ./bin/linux-amd64
	TARGET_GOOS=linux GOARCH=amd64 ./scripts/build_binary.sh ./bin/linux-amd64
	@echo "Built local-container-endpoints for linux-amd64"

$(AMD_BINARY): $(SOURCES)
	@mkdir -p ./bin/linux-arm64
	TARGET_GOOS=linux GOARCH=arm64 ./scripts/build_binary.sh ./bin/linux-arm64
	@echo "Built local-container-endpoints for linux-arm64"

.PHONY: release
release: release-amd release-arm

.PHONY: release-amd
release-amd:
	docker run -v $(shell pwd):/usr/src/app/src/github.com/awslabs/amazon-ecs-local-container-endpoints \
		--workdir=/usr/src/app/src/github.com/awslabs/amazon-ecs-local-container-endpoints \
		--env GOPATH=/usr/src/app \
		--env ECS_RELEASE=cleanbuild \
		golang:$(GO_VERSION) make $(AMD_BINARY)
	docker build -t amazon/amazon-ecs-local-container-endpoints:latest-amd64 .
	docker tag amazon/amazon-ecs-local-container-endpoints:latest amazon/amazon-ecs-local-container-endpoints:$(TAG)-amd64
	docker tag amazon/amazon-ecs-local-container-endpoints:latest amazon/amazon-ecs-local-container-endpoints:$(VERSION)-amd64

.PHONY: release-arm
release-arm:
	docker run -v $(shell pwd):/usr/src/app/src/github.com/awslabs/amazon-ecs-local-container-endpoints \
		--workdir=/usr/src/app/src/github.com/awslabs/amazon-ecs-local-container-endpoints \
		--env GOPATH=/usr/src/app \
		--env ECS_RELEASE=cleanbuild \
		golang:$(GO_VERSION) make $(ARM_BINARY)
	docker build -t amazon/amazon-ecs-local-container-endpoints:latest-arm64 .
	docker tag amazon/amazon-ecs-local-container-endpoints:latest amazon/amazon-ecs-local-container-endpoints:$(TAG)-arm64
	docker tag amazon/amazon-ecs-local-container-endpoints:latest amazon/amazon-ecs-local-container-endpoints:$(VERSION)-arm64

.PHONY: integ
integ: release
	docker build -t amazon-ecs-local-container-endpoints-integ-test:latest -f ./integ/Dockerfile .
	docker-compose --file ./integ/docker-compose.yml up --abort-on-container-exit

.PHONY: publish
publish: release publish-amd publish-arm

.PHONY: publish-amd
publish-amd:
	docker push amazon/amazon-ecs-local-container-endpoints:latest-amd64
	docker push amazon/amazon-ecs-local-container-endpoints:$(TAG)-amd64
	docker push amazon/amazon-ecs-local-container-endpoints:$(VERSION)-amd64

.PHONY: publish-arn
publish-arn:
	docker push amazon/amazon-ecs-local-container-endpoints:latest-arm64
	docker push amazon/amazon-ecs-local-container-endpoints:$(TAG)-arm64
	docker push amazon/amazon-ecs-local-container-endpoints:$(VERSION)-arm64

.PHONY: clean
clean:
	rm bin/local-container-endpoints
