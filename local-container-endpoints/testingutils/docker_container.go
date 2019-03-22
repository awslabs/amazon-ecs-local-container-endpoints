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
	"fmt"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/network"
)

const (
	image             = "ecs-local-metadata_shell"
	imageID           = "sha256:11edcbc416845013254cbab0726bb65abcc6eea1981254a888659381a630aa20"
	publicPort        = 8000
	privatePort       = 80
	protocol          = "tcp"
	networkName       = "bridge"
	ipAddress         = "172.17.0.2"
	volumeName        = "volume0"
	volumeSource      = "/var/run"
	volumeDestination = "/run"
	createdAt         = 1552368275
)

// DockerContainer wraps types.Container, and makes it easy to create
// mock responses in tests
type DockerContainer struct {
	container types.Container
}

// BaseDockerContainer returns a base container that can be customized
func BaseDockerContainer(name, containerID string) *DockerContainer {
	dockerContainer := types.Container{
		ID: containerID,
		Names: []string{
			fmt.Sprintf("/%s", name),
		},
		Image:   image,
		ImageID: imageID,
		Ports: []types.Port{
			types.Port{
				IP:          "0.0.0.0",
				PrivatePort: privatePort,
				PublicPort:  publicPort,
				Type:        protocol,
			},
		},
		Created: createdAt,
		Mounts: []types.MountPoint{
			types.MountPoint{
				Name:        volumeName,
				Source:      volumeSource,
				Destination: volumeDestination,
			},
		},
	}

	return &DockerContainer{
		container: dockerContainer,
	}
}

// WithComposeProject adds docker compose labels and returns the container for chaining
func (apiContainer *DockerContainer) WithComposeProject(projectName string) *DockerContainer {
	labels := map[string]string{
		"com.docker.compose.config-hash":      "0e48fcb738f3d237e6681f0e22f32a04172949211dee8290da691925e8ed937c",
		"com.docker.compose.container-number": "1",
		"com.docker.compose.oneoff":           "False",
		"com.docker.compose.project":          projectName,
		"com.docker.compose.service":          "ecs-local",
		"com.docker.compose.version":          "1.23.2",
	}

	apiContainer.container.Labels = labels
	return apiContainer
}

// WithNetwork adds a Docker Network and returns the container for chaining
func (apiContainer *DockerContainer) WithNetwork(networkName, ipAddress string) *DockerContainer {
	if apiContainer.container.NetworkSettings == nil {
		apiContainer.container.NetworkSettings = &types.SummaryNetworkSettings{
			Networks: make(map[string]*network.EndpointSettings),
		}
	}
	apiContainer.container.NetworkSettings.Networks[networkName] = &network.EndpointSettings{
		NetworkID: "e8884d2d5eb158e35d2d78d012e265834fb0da9cd42a288b6a5d70bfc735c84c",
		Gateway:   "172.17.0.1",
		IPAddress: ipAddress,
	}
	return apiContainer
}

// Get returns the underlying types.Container
func (apiContainer *DockerContainer) Get() types.Container {
	return apiContainer.container
}
