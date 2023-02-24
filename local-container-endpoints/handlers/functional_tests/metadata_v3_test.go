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

// package functional_tests includes tests that make http requests to the handlers using net/http/test
package functionaltests

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"strings"
	"testing"

	v2 "github.com/aws/amazon-ecs-agent/agent/handlers/v2"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/awslabs/amazon-ecs-local-container-endpoints/local-container-endpoints/clients/docker/mock_docker"
	"github.com/awslabs/amazon-ecs-local-container-endpoints/local-container-endpoints/config"
	"github.com/awslabs/amazon-ecs-local-container-endpoints/local-container-endpoints/handlers"
	"github.com/awslabs/amazon-ecs-local-container-endpoints/local-container-endpoints/testingutils"
	"github.com/docker/docker/api/types"
	"github.com/golang/mock/gomock"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

// Tests Path: /v3/containers/<container identifier>/task
func TestV3Handler_TaskMetadata(t *testing.T) {
	// Docker API Containers
	endpointsContainer := testingutils.BaseDockerContainer("endpoints", endpointsLongID).WithNetwork(network1, ipAddress).WithComposeProject(projectName).Get()
	container1 := testingutils.BaseDockerContainer(containerName1, longID1).WithNetwork(network2, ipAddress1).WithComposeProject(projectName2).Get()
	container2 := testingutils.BaseDockerContainer(containerName2, longID2).WithNetwork(network1, ipAddress2).WithComposeProject(projectName).Get()
	container3 := testingutils.BaseDockerContainer(containerName3, longID3).WithNetwork(network1, ipAddress1).WithComposeProject(projectName).Get()

	// Metadata response containers
	endpointsContainerMetadata := testingutils.BaseMetadataContainer("endpoints", endpointsLongID).WithNetwork(network1, ipAddress).WithComposeProject(projectName).Get()
	container2Metadata := testingutils.BaseMetadataContainer(containerName2, longID2).WithNetwork(network1, ipAddress2).WithComposeProject(projectName).Get()
	container3Metadata := testingutils.BaseMetadataContainer(containerName3, longID3).WithNetwork(network1, ipAddress1).WithComposeProject(projectName).Get()

	identifier := "container3"

	dockerAPIResponse := []types.Container{
		container3,
		container1,
		container2,
		endpointsContainer,
	}
	// TODO: re-enable when new task with tags metadata path is added
	// taskTags := map[string]string{
	// 	"task": "tags",
	// }
	// containerInstanceTags := map[string]string{
	// 	"containerInstance": "tags",
	// }

	os.Setenv(config.ContainerInstanceTagsVar, "containerInstance=tags")
	os.Setenv(config.TaskTagsVar, "task=tags")
	defer os.Clearenv()

	expectedMetadata := &v2.TaskResponse{
		// TaskTags:              taskTags,
		// ContainerInstanceTags: containerInstanceTags,
		Cluster:       config.DefaultClusterName,
		TaskARN:       config.DefaultTaskARN,
		Family:        config.DefaultTDFamily,
		Revision:      config.DefaultTDRevision,
		DesiredStatus: ecs.DesiredStatusRunning,
		KnownStatus:   ecs.DesiredStatusRunning,
		Containers: []v2.ContainerResponse{
			endpointsContainerMetadata,
			container2Metadata,
			container3Metadata,
		},
	}

	ctrl := gomock.NewController(t)
	dockerMock := mock_docker.NewMockClient(ctrl)

	gomock.InOrder(
		dockerMock.EXPECT().ContainerList(gomock.Any()).Return(dockerAPIResponse, nil),
	)

	metadataService, err := handlers.NewMetadataServiceWithClient(dockerMock, nil, nil)
	assert.NoError(t, err, "Unexpected error creating new metadata service")

	// create a testing server
	router := mux.NewRouter()
	metadataService.SetupV3Routes(router)
	testServer := httptest.NewServer(router)
	defer testServer.Close()

	// make a request to the testing server
	res, err := http.Get(fmt.Sprintf("%s/v3/containers/%s/task", testServer.URL, identifier))
	assert.NoError(t, err, "Unexpected error making HTTP Request")
	response, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	assert.NoError(t, err, "Unexpected error reading HTTP response")

	actualMetadata := &v2.TaskResponse{}
	err = json.Unmarshal(response, actualMetadata)
	assert.NoError(t, err, "Unexpected error unmarshalling response")

	assertContainersEqual(t, expectedMetadata.Containers, actualMetadata.Containers)
	assert.Equal(t, expectedMetadata.TaskTags, actualMetadata.TaskTags, "Expected Task Tags to match")
	assert.Equal(t, expectedMetadata.ContainerInstanceTags, actualMetadata.ContainerInstanceTags, "Expected Container Instance Tags to match")
	assert.Equal(t, expectedMetadata.Cluster, actualMetadata.Cluster, "Expected Cluster to match")
	assert.Equal(t, expectedMetadata.Family, actualMetadata.Family, "Expected Family to match")
	assert.Equal(t, expectedMetadata.Revision, actualMetadata.Revision, "Expected Revision to match")
	assert.Equal(t, expectedMetadata.DesiredStatus, actualMetadata.DesiredStatus, "Expected DesiredStatus to match")
	assert.Equal(t, expectedMetadata.KnownStatus, actualMetadata.KnownStatus, "Expected KnownStatus to match")

}

// Tests Path: /v3/containers/<container identifier>/task/
func TestV3Handler_TaskMetadata_TrailingSlash(t *testing.T) {
	// Docker API Containers
	endpointsContainer := testingutils.BaseDockerContainer("endpoints", endpointsLongID).WithNetwork(network1, ipAddress).WithComposeProject(projectName).Get()
	container1 := testingutils.BaseDockerContainer(containerName1, longID1).WithNetwork(network2, ipAddress1).WithComposeProject(projectName2).Get()
	container2 := testingutils.BaseDockerContainer(containerName2, longID2).WithNetwork(network1, ipAddress2).WithComposeProject(projectName).Get()
	container3 := testingutils.BaseDockerContainer(containerName3, longID3).WithNetwork(network1, ipAddress1).WithComposeProject(projectName).Get()

	// Metadata response containers
	endpointsContainerMetadata := testingutils.BaseMetadataContainer("endpoints", endpointsLongID).WithNetwork(network1, ipAddress).WithComposeProject(projectName).Get()
	container2Metadata := testingutils.BaseMetadataContainer(containerName2, longID2).WithNetwork(network1, ipAddress2).WithComposeProject(projectName).Get()
	container3Metadata := testingutils.BaseMetadataContainer(containerName3, longID3).WithNetwork(network1, ipAddress1).WithComposeProject(projectName).Get()

	identifier := "container3"

	dockerAPIResponse := []types.Container{
		container3,
		container1,
		container2,
		endpointsContainer,
	}
	// TODO: re-enable when new task with tags metadata path is added
	// taskTags := map[string]string{
	// 	"task": "tags",
	// }
	// containerInstanceTags := map[string]string{
	// 	"containerInstance": "tags",
	// }

	os.Setenv(config.ContainerInstanceTagsVar, "containerInstance=tags")
	os.Setenv(config.TaskTagsVar, "task=tags")
	defer os.Clearenv()

	expectedMetadata := &v2.TaskResponse{
		// TaskTags:              taskTags,
		// ContainerInstanceTags: containerInstanceTags,
		Cluster:       config.DefaultClusterName,
		TaskARN:       config.DefaultTaskARN,
		Family:        config.DefaultTDFamily,
		Revision:      config.DefaultTDRevision,
		DesiredStatus: ecs.DesiredStatusRunning,
		KnownStatus:   ecs.DesiredStatusRunning,
		Containers: []v2.ContainerResponse{
			endpointsContainerMetadata,
			container2Metadata,
			container3Metadata,
		},
	}

	ctrl := gomock.NewController(t)
	dockerMock := mock_docker.NewMockClient(ctrl)

	gomock.InOrder(
		dockerMock.EXPECT().ContainerList(gomock.Any()).Return(dockerAPIResponse, nil),
	)

	metadataService, err := handlers.NewMetadataServiceWithClient(dockerMock, nil, nil)
	assert.NoError(t, err, "Unexpected error creating new metadata service")

	// create a testing server
	router := mux.NewRouter()
	metadataService.SetupV3Routes(router)
	testServer := httptest.NewServer(router)
	defer testServer.Close()

	// make a request to the testing server
	res, err := http.Get(fmt.Sprintf("%s/v3/containers/%s/task/", testServer.URL, identifier))
	assert.NoError(t, err, "Unexpected error making HTTP Request")
	response, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	assert.NoError(t, err, "Unexpected error reading HTTP response")

	actualMetadata := &v2.TaskResponse{}
	err = json.Unmarshal(response, actualMetadata)
	assert.NoError(t, err, "Unexpected error unmarshalling response")

	assertContainersEqual(t, expectedMetadata.Containers, actualMetadata.Containers)
	assert.Equal(t, expectedMetadata.TaskTags, actualMetadata.TaskTags, "Expected Task Tags to match")
	assert.Equal(t, expectedMetadata.ContainerInstanceTags, actualMetadata.ContainerInstanceTags, "Expected Container Instance Tags to match")
	assert.Equal(t, expectedMetadata.Cluster, actualMetadata.Cluster, "Expected Cluster to match")
	assert.Equal(t, expectedMetadata.Family, actualMetadata.Family, "Expected Family to match")
	assert.Equal(t, expectedMetadata.Revision, actualMetadata.Revision, "Expected Revision to match")
	assert.Equal(t, expectedMetadata.DesiredStatus, actualMetadata.DesiredStatus, "Expected DesiredStatus to match")
	assert.Equal(t, expectedMetadata.KnownStatus, actualMetadata.KnownStatus, "Expected KnownStatus to match")

}

func TestV3Handler_TaskMetadata_DockerAPIError(t *testing.T) {
	ctrl := gomock.NewController(t)
	dockerMock := mock_docker.NewMockClient(ctrl)

	gomock.InOrder(
		dockerMock.EXPECT().ContainerList(gomock.Any()).Return(nil, fmt.Errorf("Some API Error")),
	)

	metadataService, err := handlers.NewMetadataServiceWithClient(dockerMock, nil, nil)
	assert.NoError(t, err, "Unexpected error creating new metadata service")

	// create a testing server
	router := mux.NewRouter()
	metadataService.SetupV3Routes(router)
	testServer := httptest.NewServer(router)
	defer testServer.Close()

	// make a request to the testing server
	response, err := http.Get(fmt.Sprintf("%s/v3/containers/%s/task/", testServer.URL, "container3"))
	assert.NoError(t, err, "Unexpected error making HTTP Request")
	assert.True(t, strings.Contains(response.Status, strconv.Itoa(http.StatusInternalServerError)), "Expected http response status to be internal server error")
}

func TestV3Handler_TaskMetadata_InvalidURL(t *testing.T) {
	ctrl := gomock.NewController(t)
	dockerMock := mock_docker.NewMockClient(ctrl)

	metadataService, err := handlers.NewMetadataServiceWithClient(dockerMock, nil, nil)
	assert.NoError(t, err, "Unexpected error creating new metadata service")

	// create a testing server
	router := mux.NewRouter()
	metadataService.SetupV3Routes(router)
	testServer := httptest.NewServer(router)
	defer testServer.Close()

	// make a request to the testing server
	response, err := http.Get(fmt.Sprintf("%s/v3/cats/", testServer.URL))
	assert.NoError(t, err, "Unexpected error making HTTP Request")
	assert.True(t, strings.Contains(response.Status, strconv.Itoa(http.StatusNotFound)), "Expected http response status to be 404 not found")
}

func TestV3Handler_ContainerStats(t *testing.T) {
	// Docker API Containers
	endpointsContainer := testingutils.BaseDockerContainer("endpoints", endpointsLongID).WithNetwork(network1, ipAddress).WithComposeProject(projectName).Get()
	container1 := testingutils.BaseDockerContainer(containerName1, longID1).WithNetwork(network2, ipAddress1).WithComposeProject(projectName2).Get()
	container2 := testingutils.BaseDockerContainer(containerName2, longID2).WithNetwork(network1, ipAddress2).WithComposeProject(projectName).Get()
	container3 := testingutils.BaseDockerContainer(containerName3, longID3).WithNetwork(network1, ipAddress1).WithComposeProject(projectName).Get()

	dockerAPIResponse := []types.Container{
		container3,
		container1,
		container2,
		endpointsContainer,
	}

	ctrl := gomock.NewController(t)
	dockerMock := mock_docker.NewMockClient(ctrl)

	expectedStats := getMockStats()

	gomock.InOrder(
		dockerMock.EXPECT().ContainerList(gomock.Any()).Return(dockerAPIResponse, nil),
		dockerMock.EXPECT().ContainerStats(gomock.Any(), longID1).Return(expectedStats, nil),
	)

	metadataService, err := handlers.NewMetadataServiceWithClient(dockerMock, nil, nil)
	assert.NoError(t, err, "Unexpected error creating new metadata service")

	// create a testing server
	router := mux.NewRouter()
	metadataService.SetupV3Routes(router)
	testServer := httptest.NewServer(router)
	defer testServer.Close()

	// make a request to the testing server
	res, err := http.Get(fmt.Sprintf("%s/v3/containers/%s/stats", testServer.URL, longID1))
	assert.NoError(t, err, "Unexpected error making HTTP Request")
	response, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	assert.NoError(t, err, "Unexpected error reading HTTP response")

	actualStats := &types.Stats{}
	err = json.Unmarshal(response, actualStats)
	assert.NoError(t, err, "Unexpected error unmarshalling response")

	assert.Equal(t, expectedStats, actualStats, "Expected container stats response to match")
}

func TestV3Handler_ContainerStats_TrailingSlash(t *testing.T) {
	// Docker API Containers
	endpointsContainer := testingutils.BaseDockerContainer("endpoints", endpointsLongID).WithNetwork(network1, ipAddress).WithComposeProject(projectName).Get()
	container1 := testingutils.BaseDockerContainer(containerName1, longID1).WithNetwork(network2, ipAddress1).WithComposeProject(projectName2).Get()
	container2 := testingutils.BaseDockerContainer(containerName2, longID2).WithNetwork(network1, ipAddress2).WithComposeProject(projectName).Get()
	container3 := testingutils.BaseDockerContainer(containerName3, longID3).WithNetwork(network1, ipAddress1).WithComposeProject(projectName).Get()

	dockerAPIResponse := []types.Container{
		container3,
		container1,
		container2,
		endpointsContainer,
	}

	ctrl := gomock.NewController(t)
	dockerMock := mock_docker.NewMockClient(ctrl)

	expectedStats := getMockStats()

	gomock.InOrder(
		dockerMock.EXPECT().ContainerList(gomock.Any()).Return(dockerAPIResponse, nil),
		dockerMock.EXPECT().ContainerStats(gomock.Any(), longID1).Return(expectedStats, nil),
	)

	metadataService, err := handlers.NewMetadataServiceWithClient(dockerMock, nil, nil)
	assert.NoError(t, err, "Unexpected error creating new metadata service")

	// create a testing server
	router := mux.NewRouter()
	metadataService.SetupV3Routes(router)
	testServer := httptest.NewServer(router)
	defer testServer.Close()

	// make a request to the testing server
	res, err := http.Get(fmt.Sprintf("%s/v3/containers/%s/stats", testServer.URL, longID1))
	assert.NoError(t, err, "Unexpected error making HTTP Request")
	response, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	assert.NoError(t, err, "Unexpected error reading HTTP response")

	actualStats := &types.Stats{}
	err = json.Unmarshal(response, actualStats)
	assert.NoError(t, err, "Unexpected error unmarshalling response")

	assert.Equal(t, expectedStats, actualStats, "Expected container stats response to match")
}

// Tests Path: /v2/stats/
func TestV3Handler_TaskStats(t *testing.T) {
	// Docker API Containers
	endpointsContainer := testingutils.BaseDockerContainer("endpoints", endpointsLongID).WithNetwork(network1, ipAddress).WithComposeProject(projectName).Get()
	container1 := testingutils.BaseDockerContainer(containerName1, longID1).WithNetwork(network2, ipAddress1).WithComposeProject(projectName2).Get()
	container2 := testingutils.BaseDockerContainer(containerName2, longID2).WithNetwork(network1, ipAddress2).WithComposeProject(projectName).Get()
	container3 := testingutils.BaseDockerContainer(containerName3, longID3).WithNetwork(network1, ipAddress1).WithComposeProject(projectName).Get()

	dockerAPIResponse := []types.Container{
		container3,
		container1,
		container2,
		endpointsContainer,
	}

	ctrl := gomock.NewController(t)
	dockerMock := mock_docker.NewMockClient(ctrl)

	container1Stats := getMockStats()
	container2Stats := getMockStats()
	container3Stats := getMockStats()
	endpointsStats := getMockStats()

	expectedStats := map[string]types.Stats{
		longID1:         *container1Stats,
		longID2:         *container2Stats,
		longID3:         *container3Stats,
		endpointsLongID: *endpointsStats,
	}

	dockerMock.EXPECT().ContainerList(gomock.Any()).Return(dockerAPIResponse, nil)
	dockerMock.EXPECT().ContainerStats(gomock.Any(), longID1).Return(container1Stats, nil)
	dockerMock.EXPECT().ContainerStats(gomock.Any(), longID2).Return(container2Stats, nil)
	dockerMock.EXPECT().ContainerStats(gomock.Any(), longID3).Return(container3Stats, nil)
	dockerMock.EXPECT().ContainerStats(gomock.Any(), endpointsLongID).Return(endpointsStats, nil)

	metadataService, err := handlers.NewMetadataServiceWithClient(dockerMock, nil, nil)
	assert.NoError(t, err, "Unexpected error creating new metadata service")

	// create a testing server
	router := mux.NewRouter()
	metadataService.SetupV3Routes(router)
	testServer := httptest.NewServer(router)
	defer testServer.Close()

	// make a request to the testing server
	res, err := http.Get(fmt.Sprintf("%s/v3/task/stats", testServer.URL))
	assert.NoError(t, err, "Unexpected error making HTTP Request")
	response, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	assert.NoError(t, err, "Unexpected error reading HTTP response")

	actualStats := make(map[string]types.Stats)
	err = json.Unmarshal(response, &actualStats)
	assert.NoError(t, err, "Unexpected error unmarshalling response")

	assert.Equal(t, expectedStats, actualStats, "Expected container stats response to match")
}

func TestV3Handler_TaskStats_TrailingSlash(t *testing.T) {
	// Docker API Containers
	endpointsContainer := testingutils.BaseDockerContainer("endpoints", endpointsLongID).WithNetwork(network1, ipAddress).WithComposeProject(projectName).Get()
	container1 := testingutils.BaseDockerContainer(containerName1, longID1).WithNetwork(network2, ipAddress1).WithComposeProject(projectName2).Get()
	container2 := testingutils.BaseDockerContainer(containerName2, longID2).WithNetwork(network1, ipAddress2).WithComposeProject(projectName).Get()
	container3 := testingutils.BaseDockerContainer(containerName3, longID3).WithNetwork(network1, ipAddress1).WithComposeProject(projectName).Get()

	dockerAPIResponse := []types.Container{
		container3,
		container1,
		container2,
		endpointsContainer,
	}

	ctrl := gomock.NewController(t)
	dockerMock := mock_docker.NewMockClient(ctrl)

	container1Stats := getMockStats()
	container2Stats := getMockStats()
	container3Stats := getMockStats()
	endpointsStats := getMockStats()

	expectedStats := map[string]types.Stats{
		longID1:         *container1Stats,
		longID2:         *container2Stats,
		longID3:         *container3Stats,
		endpointsLongID: *endpointsStats,
	}

	dockerMock.EXPECT().ContainerList(gomock.Any()).Return(dockerAPIResponse, nil)
	dockerMock.EXPECT().ContainerStats(gomock.Any(), longID1).Return(container1Stats, nil)
	dockerMock.EXPECT().ContainerStats(gomock.Any(), longID2).Return(container2Stats, nil)
	dockerMock.EXPECT().ContainerStats(gomock.Any(), longID3).Return(container3Stats, nil)
	dockerMock.EXPECT().ContainerStats(gomock.Any(), endpointsLongID).Return(endpointsStats, nil)

	metadataService, err := handlers.NewMetadataServiceWithClient(dockerMock, nil, nil)
	assert.NoError(t, err, "Unexpected error creating new metadata service")

	// create a testing server
	router := mux.NewRouter()
	metadataService.SetupV3Routes(router)
	testServer := httptest.NewServer(router)
	defer testServer.Close()

	// make a request to the testing server
	res, err := http.Get(fmt.Sprintf("%s/v3/task/stats/", testServer.URL))
	assert.NoError(t, err, "Unexpected error making HTTP Request")
	response, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	assert.NoError(t, err, "Unexpected error reading HTTP response")

	actualStats := make(map[string]types.Stats)
	err = json.Unmarshal(response, &actualStats)
	assert.NoError(t, err, "Unexpected error unmarshalling response")

	assert.Equal(t, expectedStats, actualStats, "Expected container stats response to match")
}

func TestV3Handler_TaskStats_DockerAPIError(t *testing.T) {
	// Docker API Containers
	endpointsContainer := testingutils.BaseDockerContainer("endpoints", endpointsLongID).WithNetwork(network1, ipAddress).WithComposeProject(projectName).Get()
	container1 := testingutils.BaseDockerContainer(containerName1, longID1).WithNetwork(network2, ipAddress1).WithComposeProject(projectName2).Get()
	container2 := testingutils.BaseDockerContainer(containerName2, longID2).WithNetwork(network1, ipAddress2).WithComposeProject(projectName).Get()
	container3 := testingutils.BaseDockerContainer(containerName3, longID3).WithNetwork(network1, ipAddress1).WithComposeProject(projectName).Get()

	dockerAPIResponse := []types.Container{
		container3,
		container1,
		container2,
		endpointsContainer,
	}

	ctrl := gomock.NewController(t)
	dockerMock := mock_docker.NewMockClient(ctrl)

	container1Stats := getMockStats()
	container2Stats := getMockStats()
	endpointsStats := getMockStats()

	dockerMock.EXPECT().ContainerList(gomock.Any()).Return(dockerAPIResponse, nil)
	dockerMock.EXPECT().ContainerStats(gomock.Any(), longID1).Return(container1Stats, nil)
	dockerMock.EXPECT().ContainerStats(gomock.Any(), longID2).Return(container2Stats, nil)
	dockerMock.EXPECT().ContainerStats(gomock.Any(), longID3).Return(nil, fmt.Errorf("Some error"))
	dockerMock.EXPECT().ContainerStats(gomock.Any(), endpointsLongID).Return(endpointsStats, nil)

	metadataService, err := handlers.NewMetadataServiceWithClient(dockerMock, nil, nil)
	assert.NoError(t, err, "Unexpected error creating new metadata service")

	// create a testing server
	router := mux.NewRouter()
	metadataService.SetupV3Routes(router)
	testServer := httptest.NewServer(router)
	defer testServer.Close()

	// make a request to the testing server
	response, err := http.Get(fmt.Sprintf("%s/v3/task/stats", testServer.URL))
	assert.NoError(t, err, "Unexpected error making HTTP Request")
	assert.True(t, strings.Contains(response.Status, strconv.Itoa(http.StatusInternalServerError)), "Expected http response status to be internal server error")
}
