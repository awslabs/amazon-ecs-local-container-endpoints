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
	"strings"
	"time"

	"github.com/aws/amazon-ecs-agent/agent/containermetadata"
	"github.com/aws/amazon-ecs-agent/agent/handlers/v1"
	"github.com/aws/amazon-ecs-agent/agent/handlers/v2"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/awslabs/amazon-ecs-local-container-endpoints/modules/config"
	"github.com/awslabs/amazon-ecs-local-container-endpoints/modules/utils"
	"github.com/docker/docker/api/types"
)

// GetTaskMetadata returns the task metadata for the given containers
func GetTaskMetadata(dockerContainers []types.Container, containerInstanceTags, taskTags map[string]string) *v2.TaskResponse {
	response := newMockTaskResponse(containerInstanceTags, taskTags)
	ecsContainers := response.Containers
	for _, container := range dockerContainers {
		ecsContainer := GetContainerMetadata(&container)
		ecsContainers = append(ecsContainers, *ecsContainer)
	}
	response.Containers = ecsContainers
	return response
}

// GetContainerMetadata creates a container metadata response using info from the docker API,
// with other values mocked
func GetContainerMetadata(dockerContainer *types.Container) *v2.ContainerResponse {
	response := newMockContainerResponse()
	response.ID = dockerContainer.ID
	response.Name = getContainerName(dockerContainer)
	response.DockerName = getContainerName(dockerContainer)
	response.Image = dockerContainer.Image
	response.ImageID = dockerContainer.ImageID
	response.Ports = convertPorts(dockerContainer.Ports)
	response.Labels = dockerContainer.Labels
	createTime := time.Unix(dockerContainer.Created, 0)
	response.CreatedAt = &createTime
	// we can't know the actual start time, but we err on the side of having as many values in the response as possible
	response.StartedAt = response.CreatedAt
	response.Networks = convertNetworks(dockerContainer.NetworkSettings)
	response.Volumes = convertVolumes(dockerContainer.Mounts)

	return response
}

func newMockContainerResponse() *v2.ContainerResponse {
	return &v2.ContainerResponse{
		DesiredStatus: ecs.DesiredStatusRunning,
		KnownStatus:   ecs.DesiredStatusRunning,
		Type:          config.DefaultContainerType,
	}
}

func newMockTaskResponse(containerInstanceTags, taskTags map[string]string) *v2.TaskResponse {
	return &v2.TaskResponse{
		Cluster:               utils.GetValue(config.DefaultClusterName, config.ClusterARNVar),
		TaskARN:               utils.GetValue(config.DefaultTaskARN, config.TaskARNVar),
		Family:                utils.GetValue(config.DefaultTDFamily, config.TDFamilyVar),
		Revision:              utils.GetValue(config.DefaultTDRevision, config.TDRevisionVar),
		DesiredStatus:         ecs.DesiredStatusRunning,
		KnownStatus:           ecs.DesiredStatusRunning,
		TaskTags:              taskTags,
		ContainerInstanceTags: containerInstanceTags,
	}
}

func convertVolumes(mounts []types.MountPoint) []v1.VolumeResponse {
	ecsVolumes := make([]v1.VolumeResponse, len(mounts))
	for _, mount := range mounts {
		ecsVolumes = append(ecsVolumes, v1.VolumeResponse{
			DockerName:  mount.Name,
			Source:      mount.Source,
			Destination: mount.Destination,
		})
	}
	return ecsVolumes
}

func convertNetworks(dockerNetworkSettings *types.SummaryNetworkSettings) []containermetadata.Network {
	ecsNetworks := make([]containermetadata.Network, len(dockerNetworkSettings.Networks))
	for netMode, netSettings := range dockerNetworkSettings.Networks {
		ecsNet := containermetadata.Network{
			NetworkMode: netMode,
		}
		if netSettings.IPAddress != "" {
			ecsNet.IPv4Addresses = []string{
				netSettings.IPAddress,
			}
		}
		if netSettings.GlobalIPv6Address != "" {
			ecsNet.IPv6Addresses = []string{
				netSettings.GlobalIPv6Address,
			}
		}
		ecsNetworks = append(ecsNetworks, ecsNet)
	}
	return ecsNetworks
}

func convertPorts(dockerPorts []types.Port) []v1.PortResponse {
	ecsPorts := make([]v1.PortResponse, len(dockerPorts))
	for _, port := range dockerPorts {
		ecsPorts = append(ecsPorts, v1.PortResponse{
			ContainerPort: port.PrivatePort,
			HostPort:      port.PublicPort,
			Protocol:      port.Type,
		})
	}
	return ecsPorts
}

// Docker API returns a list of container names, each prefixed by a slash
// This function returns the first name in the list, and removes the slash (which is not present in the ECS Metadata response)
func getContainerName(dockerContainer *types.Container) string {
	if len(dockerContainer.Names) > 0 {
		return strings.Trim(dockerContainer.Names[0], "/")
	}
	return ""
}
