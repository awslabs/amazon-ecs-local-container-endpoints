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
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/network"
	"github.com/stretchr/testify/assert"
)

const (
	cluster   = "meow-cluster"
	taskARN   = "arn:aws-cats:ecs:us-west-2:111111111111:task/meow-cluster/37e873f6-37b4-42a7-af47-eac7275c6152"
	family    = "the-internet-is-for-cats"
	revision  = "2"
	createdAt = 1552368275
)

func TestNewMockTaskResponseWithEnvVars(t *testing.T) {
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

	actual := newMockTaskResponse(nil, nil)
	assert.Equal(t, expected, actual, "Expected TaskResponse to match")
}

func TestGetTaskMetadata(t *testing.T) {
	dockerContainer := types.Container{
		ID: " c3439823c17dc7a35c7e272b7dc51cb2dcdedcef428242fcd0f5473d2c724d0",
		Names: []string{
			"/ecs-local-metadata_shell_1",
		},
		Image:   "ecs-local-metadata_shell",
		ImageID: "sha256:11edcbc416845013254cbab0726bb65abcc6eea1981254a888659381a630aa20",
		Ports: []types.Port{
			types.Port{
				IP:          "0.0.0.0",
				PrivatePort: 80,
				PublicPort:  80,
				Type:        "tcp",
			},
		},
		Labels: map[string]string{
			"com.docker.compose.config-hash":      "0e48fcb738f3d237e6681f0e22f32a04172949211dee8290da691925e8ed937c",
			"com.docker.compose.container-number": "1",
			"com.docker.compose.oneoff":           "False",
			"com.docker.compose.project":          "ecs-local-metadata",
			"com.docker.compose.service":          "ecs-local",
			"com.docker.compose.version":          "1.23.2",
		},
		Created: createdAt,
		NetworkSettings: &types.SummaryNetworkSettings{
			Networks: map[string]*network.EndpointSettings{
				"bridge": &network.EndpointSettings{
					NetworkID: "e8884d2d5eb158e35d2d78d012e265834fb0da9cd42a288b6a5d70bfc735c84c",
					Gateway:   "172.17.0.1",
					IPAddress: "172.17.0.2",
				},
			},
		},
		Mounts: []types.MountPoint{
			types.MountPoint{
				Name:        "volume0",
				Source:      "/var/run",
				Destination: "/run",
			},
		},
	}

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
			v2.ContainerResponse{
				DesiredStatus: ecs.DesiredStatusRunning,
				KnownStatus:   ecs.DesiredStatusRunning,
				Type:          config.DefaultContainerType,
			},
		},
	}
	//GetTaskMetadata([], containerInstanceTags, taskTags)

}
