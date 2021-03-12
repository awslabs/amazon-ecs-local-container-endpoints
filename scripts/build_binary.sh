#!/bin/bash
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

# Normalize to working directory being build root (up one level from ./scripts)
ROOT=$( cd "$( dirname "${BASH_SOURCE[0]}" )/.." && pwd )
cd "${ROOT}"

# Builds the binary from source in the specified destination paths.
mkdir -p $1

# Versioning stuff. We run the generator to setup the version and then always
# restore ourselves to a clean state
cp local-container-endpoints/version/version.go local-container-endpoints/version/_version.go
trap "cd \"${ROOT}\"; mv local-container-endpoints/version/_version.go local-container-endpoints/version/version.go" EXIT SIGHUP SIGINT SIGTERM

cd local-container-endpoints/version/
go run gen/version-gen.go

cd "${ROOT}"

GOOS=$TARGET_GOOS GOARCH=$GOARCH CGO_ENABLED=0 GO111MODULE=on go build -mod=vendor -installsuffix cgo -a -ldflags '-s' -o $1/local-container-endpoints ./
