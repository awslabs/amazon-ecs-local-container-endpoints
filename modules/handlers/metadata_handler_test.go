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

package handlers

import (
	"os"
	"testing"

	"github.com/awslabs/amazon-ecs-local-container-endpoints/modules/config"
	"github.com/docker/docker/api/types"
	"github.com/stretchr/testify/assert"
)

const (
	shortID     = "56771b9219b5"
	longID      = "56771b9219b58c8b6a286830667b62475e79753db34a0b82a98efafb20718c0f9"
	projectName = "project"
)

func containerInProject(longID string) types.Container {
	return types.Container{
		ID: longID,
		Labels: map[string]string{
			composeProjectNameLabel: projectName,
		},
	}
}

func containerOutsideProject(longID string) types.Container {
	return types.Container{
		ID: longID,
	}
}

func TestFilterByComposeProject(t *testing.T) {
	endpointsContainer := types.Container{
		ID: longID,
		Labels: map[string]string{
			composeProjectNameLabel: projectName,
		},
	}
	containerInCompose := containerInProject("e18ab3d25b38c8b6a286830667b62475e79753db34a0b82a98efafb20718c0f9")
	containerOutsideCompose := containerOutsideProject("sadwes23084b6a286830667b62475e79753db3sdlk28932036efafb20718cf09")

	inputContainers := []types.Container{
		endpointsContainer,
		containerInCompose,
		containerOutsideCompose,
	}
	expectedContainers := []types.Container{
		endpointsContainer,
		containerInCompose,
	}

	os.Setenv("HOSTNAME", shortID)
	defer os.Setenv("HOSTNAME", "")

	observed := filterByComposeProject(inputContainers)
	assert.ElementsMatch(t, observed, expectedContainers, "Expected containers to match after filtering")
}

// Tests the case where the Endpoints Container is running in compose, but no other containers are
// In this case, we want to return metadata for all running containers
func TestFilterByComposeProjectNoOtherContainersInCompose(t *testing.T) {
	endpointsContainer := types.Container{
		ID: longID,
		Labels: map[string]string{
			composeProjectNameLabel: projectName,
		},
	}
	containerOutsideCompose := containerOutsideProject("e18ab3d25b38c8b6a286830667b62475e79753db34a0b82a98efafb20718c0f9")
	containerOutsideCompose2 := containerOutsideProject("sadwes23084b6a286830667b62475e79753db3sdlk28932036efafb20718cf09")

	inputContainers := []types.Container{
		endpointsContainer,
		containerOutsideCompose,
		containerOutsideCompose2,
	}
	expectedContainers := []types.Container{
		endpointsContainer,
		containerOutsideCompose,
		containerOutsideCompose2,
	}

	os.Setenv("HOSTNAME", shortID)
	defer os.Setenv("HOSTNAME", "")

	observed := filterByComposeProject(inputContainers)
	assert.ElementsMatch(t, observed, expectedContainers, "Expected containers to match after filtering")
}

// Tests the case where the endpoints container is not in a compose project
func TestFilterByComposeProjectNotInCompose(t *testing.T) {
	endpointsContainer := types.Container{
		ID: longID,
	}
	containerInCompose := containerInProject("e18ab3d25b38c8b6a286830667b62475e79753db34a0b82a98efafb20718c0f9")
	containerOutsideCompose := containerOutsideProject("sadwes23084b6a286830667b62475e79753db3sdlk28932036efafb20718cf09")

	inputContainers := []types.Container{
		endpointsContainer,
		containerInCompose,
		containerOutsideCompose,
	}
	expectedContainers := []types.Container{
		endpointsContainer,
		containerInCompose,
		containerOutsideCompose,
	}

	os.Setenv("HOSTNAME", shortID)
	defer os.Setenv("HOSTNAME", "")

	observed := filterByComposeProject(inputContainers)
	assert.ElementsMatch(t, observed, expectedContainers, "Expected containers to match after filtering")
}

func TestNewMetadataServiceWithTags(t *testing.T) {
	os.Setenv(config.ContainerInstanceTagsVar, "mitchell=webb,thats=numberwang")
	os.Setenv(config.TaskTagsVar, "hello=goodbye,get=back,come=together")
	expectedCITags := map[string]string{
		"mitchell": "webb",
		"thats":    "numberwang",
	}
	expectedTaskTags := map[string]string{
		"hello": "goodbye",
		"get":   "back",
		"come":  "together",
	}

	service, err := NewMetadataService()
	assert.NoError(t, err, "Unexpected error calling NewMetadataService")
	assert.Equal(t, expectedCITags, service.containerInstanceTags, "Expected container instance tags to match")
	assert.Equal(t, expectedTaskTags, service.taskTags, "Expected task tags to match")
}
