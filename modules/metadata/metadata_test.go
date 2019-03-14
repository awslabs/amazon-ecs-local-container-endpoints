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

package metadata

import (
	"os"
	"testing"

	"github.com/aws/amazon-ecs-agent/agent/handlers/v2"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/awslabs/amazon-ecs-local-container-endpoints/modules/config"
	"github.com/awslabs/amazon-ecs-local-container-endpoints/modules/testingutils"
	"github.com/docker/docker/api/types"
	"github.com/stretchr/testify/assert"
)

const (
	cluster       = "meow-cluster"
	taskARN       = "arn:aws-cats:ecs:us-west-2:111111111111:task/meow-cluster/37e873f6-37b4-42a7-af47-eac7275c6152"
	family        = "the-internet-is-for-cats"
	revision      = "2"
	projectName   = "meow-zedong"
	ipAddress     = "127.0.0.5"
	containerID   = "c3439823c17dc7a35c7e272b7dc51cb2dcdedcef428242fcd0f5473d2c724d0"
	containerName = "ecs-local-endpoints"
)

func TestnewLocalTaskResponseWithEnvVars(t *testing.T) {
	expected := &v2.TaskResponse{
		Cluster:       cluster,
		TaskARN:       taskARN,
		Family:        family,
		Revision:      revision,
		DesiredStatus: ecs.DesiredStatusRunning,
		KnownStatus:   ecs.DesiredStatusRunning,
	}

	os.Setenv(config.ClusterARNVar, cluster)
	os.Setenv(config.TaskARNVar, taskARN)
	os.Setenv(config.TDFamilyVar, family)
	os.Setenv(config.TDRevisionVar, revision)
	defer os.Clearenv()

	actual := newLocalTaskResponse(nil, nil)
	assert.Equal(t, expected, actual, "Expected TaskResponse to match")
}

func TestGetTaskMetadata(t *testing.T) {
	dockerContainer := testingutils.BaseDockerContainer(containerName, containerID).
		WithComposeProject(projectName).
		WithNetwork("bridge", ipAddress).
		Get()

	expectedContainer := testingutils.BaseMetadataContainer(containerName, containerID).
		WithComposeProject(projectName).
		WithNetwork("bridge", ipAddress).
		Get()

	taskTags := map[string]string{
		"task": "tags",
	}
	containerInstanceTags := map[string]string{
		"containerInstance": "tags",
	}

	expected := &v2.TaskResponse{
		TaskTags:              taskTags,
		ContainerInstanceTags: containerInstanceTags,
		Cluster:               config.DefaultClusterName,
		TaskARN:               config.DefaultTaskARN,
		Family:                config.DefaultTDFamily,
		Revision:              config.DefaultTDRevision,
		DesiredStatus:         ecs.DesiredStatusRunning,
		KnownStatus:           ecs.DesiredStatusRunning,
		Containers: []v2.ContainerResponse{
			expectedContainer,
		},
	}

	actual := GetTaskMetadata([]types.Container{dockerContainer}, containerInstanceTags, taskTags)
	assert.Equal(t, expected, actual, "Expected task response to match")
}
