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

// Package testingutils provides functionality that is useful in tests accross this project
package testingutils

import (
	"time"

	"github.com/aws/amazon-ecs-agent/agent/containermetadata"
	"github.com/aws/amazon-ecs-agent/agent/handlers/v1"
	"github.com/aws/amazon-ecs-agent/agent/handlers/v2"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/awslabs/amazon-ecs-local-container-endpoints/modules/config"
)

// MetadataContainer wraps v2.ContainerResponse, and makes it easy to create
// mock responses in tests
type MetadataContainer struct {
	container v2.ContainerResponse
}

// BaseMetadataContainer returns a base container that can be customized
func BaseMetadataContainer(name, containerID string) *MetadataContainer {
	createTime := time.Unix(createdAt, 0)
	container := v2.ContainerResponse{
		DesiredStatus: ecs.DesiredStatusRunning,
		KnownStatus:   ecs.DesiredStatusRunning,
		Type:          config.DefaultContainerType,
		ID:            containerID,
		Name:          name,
		DockerName:    name,
		Image:         image,
		ImageID:       imageID,
		Ports: []v1.PortResponse{
			v1.PortResponse{
				ContainerPort: privatePort,
				HostPort:      publicPort,
				Protocol:      protocol,
			},
		},
		CreatedAt: &createTime,
		StartedAt: &createTime,
		Volumes: []v1.VolumeResponse{
			v1.VolumeResponse{
				DockerName:  volumeName,
				Source:      volumeSource,
				Destination: volumeDestination,
			},
		},
	}

	return &MetadataContainer{
		container: container,
	}
}

// WithComposeProject adds docker compose labels and returns the container for chaining
func (c *MetadataContainer) WithComposeProject(projectName string) *MetadataContainer {
	labels := map[string]string{
		"com.docker.compose.config-hash":      "0e48fcb738f3d237e6681f0e22f32a04172949211dee8290da691925e8ed937c",
		"com.docker.compose.container-number": "1",
		"com.docker.compose.oneoff":           "False",
		"com.docker.compose.project":          projectName,
		"com.docker.compose.service":          "ecs-local",
		"com.docker.compose.version":          "1.23.2",
	}

	c.container.Labels = labels
	return c
}

// WithNetwork adds a Docker Network and returns the container for chaining
func (c *MetadataContainer) WithNetwork(networkName, ipAddress string) *MetadataContainer {
	c.container.Networks = append(c.container.Networks, containermetadata.Network{
		NetworkMode: networkName,
		IPv4Addresses: []string{
			ipAddress,
		},
	})
	return c
}

// Get returns the underlying v2.ContainerResponse
func (c *MetadataContainer) Get() v2.ContainerResponse {
	return c.container
}
