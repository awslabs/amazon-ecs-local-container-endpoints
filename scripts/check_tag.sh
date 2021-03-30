#!/bin/bash
# Copyright 2021 Amazon.com, Inc. or its affiliates. All Rights Reserved.
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
# ROOT=$( cd "$( dirname "${BASH_SOURCE[0]}" )/.." && pwd )
# cd "${ROOT}"

export VERSION=`cat VERSION`
export GIT_TAG=`git tag --points-at HEAD`
if [ "v${VERSION}" != $GIT_TAG ]; then
	exit 1
else
	echo "Continuing build for $GIT_TAG"
fi
