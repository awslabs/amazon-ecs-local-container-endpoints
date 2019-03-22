// Copyright 2019 Amazon.com, Inc. or its affiliates. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License"). You may
// not use this file except in compliance with the License. A copy of the
// License is located at
//
//	http://aws.amazon.com/apache2.0/
//
// or in the "license" file accompanying this file. This file is distributed
// on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either
// express or implied. See the License for the specific language governing
// permissions and limitations under the License.

package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"testing"

	"github.com/aws/amazon-ecs-agent/agent/handlers/v2"
	"github.com/stretchr/testify/assert"
)

func TestV3Handler_TaskMetadata(t *testing.T) {
	v3Path := os.Getenv("ECS_CONTAINER_METADATA_URI")
	res, err := http.Get(fmt.Sprintf("%s/task", v3Path))
	assert.NoError(t, err, "Unexpected error making HTTP Request")
	response, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	assert.NoError(t, err, "Unexpected error reading HTTP response")

	actualMetadata := &v2.TaskResponse{}
	err = json.Unmarshal(response, actualMetadata)
	assert.NoError(t, err, "Unexpected error unmarshalling response")

	assert.Len(t, actualMetadata.Containers, 3, "Expected 3 containers in response")

	expectedNames := []string{
		"integ_ecs-local_1",
		"integ_integration-test_1",
		"integ_nginx_1",
	}

	actualNames := []string{
		actualMetadata.Containers[0].Name,
		actualMetadata.Containers[1].Name,
		actualMetadata.Containers[2].Name,
	}

	assert.ElementsMatch(t, expectedNames, actualNames, "Expected list of container names to match")
}
