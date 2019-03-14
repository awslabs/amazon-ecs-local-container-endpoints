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
package functional_tests

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/aws/amazon-ecs-agent/agent/handlers/v2"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/awslabs/amazon-ecs-local-container-endpoints/modules/clients/docker/mock_docker"
	"github.com/awslabs/amazon-ecs-local-container-endpoints/modules/config"
	"github.com/awslabs/amazon-ecs-local-container-endpoints/modules/handlers"
	"github.com/awslabs/amazon-ecs-local-container-endpoints/modules/testingutils"
	"github.com/docker/docker/api/types"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

const (
	endpointsShortID = "56771b9219b5"
	endpointsLongID  = "56771b9219b58c8b6a286830667b62475e79753db34a0b82a98efafb20718c0f9"
	shortID1         = "e18ab3d25b38"
	longID1          = "e18ab3d25b38c8b6a287831767b62475a79853dc38a0b92a98efabb20718c0d90"
	longID2          = "457129ed3bd03f1fc70125c3be7bcbee760d5edf092e32155a5c6a730cd32020"
	longID3          = "0756a2371cad1976b07954490660f07d240a6a6f52d17594ed691799915695f7"
	containerName1   = "container1"
	containerName2   = "container2-pudding"
	containerName3   = "clyde-container3-dumpling"
	badName          = "tum-tum"
	ipAddress        = "169.254.170.2"
	ipAddress1       = "172.17.0.2"
	ipAddress2       = "172.17.0.3"
	ipAddress3       = "172.17.0.4"
	network1         = "metadata-network"
	network2         = "app-network"
	projectName      = "project"
	projectName2     = "operation-clyde-undercover"
)

// Tests Path: /v2/metadata
func TestV2Handler_TaskMetadata(t *testing.T) {
	// Docker API Containers
	endpointsContainer := testingutils.BaseDockerContainer("endpoints", endpointsLongID).WithNetwork(network1, ipAddress).WithComposeProject(projectName).Get()
	container1 := testingutils.BaseDockerContainer(containerName1, longID1).WithNetwork(network2, ipAddress1).WithComposeProject(projectName2).Get()
	container2 := testingutils.BaseDockerContainer(containerName2, longID2).WithNetwork(network1, ipAddress2).WithComposeProject(projectName).Get()
	container3 := testingutils.BaseDockerContainer(containerName3, longID3).WithNetwork(network1, ipAddress1).WithComposeProject(projectName).Get()

	// Metadata response containers
	endpointsContainerMetadata := testingutils.BaseMetadataContainer("endpoints", endpointsLongID).WithNetwork(network1, ipAddress).WithComposeProject(projectName).Get()
	container1Metadata := testingutils.BaseMetadataContainer(containerName1, longID1).WithNetwork(network2, ipAddress1).WithComposeProject(projectName2).Get()
	container2Metadata := testingutils.BaseMetadataContainer(containerName2, longID2).WithNetwork(network1, ipAddress2).WithComposeProject(projectName).Get()
	container3Metadata := testingutils.BaseMetadataContainer(containerName3, longID3).WithNetwork(network1, ipAddress1).WithComposeProject(projectName).Get()

	dockerAPIResponse := []types.Container{
		container3,
		container1,
		container2,
		endpointsContainer,
	}

	taskTags := map[string]string{
		"task": "tags",
	}
	containerInstanceTags := map[string]string{
		"containerInstance": "tags",
	}

	os.Setenv(config.ContainerInstanceTagsVar, "containerInstance=tags")
	os.Setenv(config.TaskTagsVar, "task=tags")
	defer os.Clearenv()

	expectedMetadata := &v2.TaskResponse{
		TaskTags:              taskTags,
		ContainerInstanceTags: containerInstanceTags,
		Cluster:               config.DefaultClusterName,
		TaskARN:               config.DefaultTaskARN,
		Family:                config.DefaultTDFamily,
		Revision:              config.DefaultTDRevision,
		DesiredStatus:         ecs.DesiredStatusRunning,
		KnownStatus:           ecs.DesiredStatusRunning,
		Containers: []v2.ContainerResponse{
			endpointsContainerMetadata,
			container1Metadata,
			container2Metadata,
			container3Metadata,
		},
	}

	ctrl := gomock.NewController(t)
	dockerMock := mock_docker.NewMockClient(ctrl)

	gomock.InOrder(
		dockerMock.EXPECT().ContainerList(gomock.Any()).Return(dockerAPIResponse, nil),
	)

	metadataService, err := handlers.NewMetadataServiceWithClient(dockerMock)
	assert.NoError(t, err, "Unexpected error creating new metadata service")

	// create a testing server
	testServer := httptest.NewServer(http.HandlerFunc(handlers.ServeHTTP(metadataService.GetV2Handler())))
	defer testServer.Close()

	// make a request to the testing server
	res, err := http.Get(fmt.Sprintf("%s/v2/metadata", testServer.URL))
	assert.NoError(t, err, "Unexpected error making HTTP Request")
	response, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	assert.NoError(t, err, "Unexpected error reading HTTP response")

	actualMetadata := &v2.TaskResponse{}
	err = json.Unmarshal(response, actualMetadata)
	assert.NoError(t, err, "Unexpected error unmarshalling response")

	assert.ElementsMatch(t, expectedMetadata.Containers, actualMetadata.Containers, "Expected container responses to match")
	assert.Equal(t, expectedMetadata.TaskTags, actualMetadata.TaskTags, "Expected Task Tags to match")
	assert.Equal(t, expectedMetadata.ContainerInstanceTags, actualMetadata.ContainerInstanceTags, "Expected Container Instance Tags to match")
	assert.Equal(t, expectedMetadata.Cluster, actualMetadata.Cluster, "Expected Cluster to match")
	assert.Equal(t, expectedMetadata.Family, actualMetadata.Family, "Expected Family to match")
	assert.Equal(t, expectedMetadata.Revision, actualMetadata.Revision, "Expected Revision to match")
	assert.Equal(t, expectedMetadata.DesiredStatus, actualMetadata.DesiredStatus, "Expected DesiredStatus to match")
	assert.Equal(t, expectedMetadata.KnownStatus, actualMetadata.KnownStatus, "Expected KnownStatus to match")

}

// Tests Path: /v2/metadata/
func TestV2Handler_TaskMetadata_TrailingSlash(t *testing.T) {
	// Docker API Containers
	endpointsContainer := testingutils.BaseDockerContainer("endpoints", endpointsLongID).WithNetwork(network1, ipAddress).WithComposeProject(projectName).Get()
	container1 := testingutils.BaseDockerContainer(containerName1, longID1).WithNetwork(network2, ipAddress1).WithComposeProject(projectName2).Get()
	container2 := testingutils.BaseDockerContainer(containerName2, longID2).WithNetwork(network1, ipAddress2).WithComposeProject(projectName).Get()
	container3 := testingutils.BaseDockerContainer(containerName3, longID3).WithNetwork(network1, ipAddress1).WithComposeProject(projectName).Get()

	// Metadata response containers
	endpointsContainerMetadata := testingutils.BaseMetadataContainer("endpoints", endpointsLongID).WithNetwork(network1, ipAddress).WithComposeProject(projectName).Get()
	container1Metadata := testingutils.BaseMetadataContainer(containerName1, longID1).WithNetwork(network2, ipAddress1).WithComposeProject(projectName2).Get()
	container2Metadata := testingutils.BaseMetadataContainer(containerName2, longID2).WithNetwork(network1, ipAddress2).WithComposeProject(projectName).Get()
	container3Metadata := testingutils.BaseMetadataContainer(containerName3, longID3).WithNetwork(network1, ipAddress1).WithComposeProject(projectName).Get()

	dockerAPIResponse := []types.Container{
		container3,
		container1,
		container2,
		endpointsContainer,
	}

	taskTags := map[string]string{
		"task": "tags",
	}
	containerInstanceTags := map[string]string{
		"containerInstance": "tags",
	}

	os.Setenv(config.ContainerInstanceTagsVar, "containerInstance=tags")
	os.Setenv(config.TaskTagsVar, "task=tags")
	defer os.Clearenv()

	expectedMetadata := &v2.TaskResponse{
		TaskTags:              taskTags,
		ContainerInstanceTags: containerInstanceTags,
		Cluster:               config.DefaultClusterName,
		TaskARN:               config.DefaultTaskARN,
		Family:                config.DefaultTDFamily,
		Revision:              config.DefaultTDRevision,
		DesiredStatus:         ecs.DesiredStatusRunning,
		KnownStatus:           ecs.DesiredStatusRunning,
		Containers: []v2.ContainerResponse{
			endpointsContainerMetadata,
			container1Metadata,
			container2Metadata,
			container3Metadata,
		},
	}

	ctrl := gomock.NewController(t)
	dockerMock := mock_docker.NewMockClient(ctrl)

	gomock.InOrder(
		dockerMock.EXPECT().ContainerList(gomock.Any()).Return(dockerAPIResponse, nil),
	)

	metadataService, err := handlers.NewMetadataServiceWithClient(dockerMock)
	assert.NoError(t, err, "Unexpected error creating new metadata service")

	// create a testing server
	testServer := httptest.NewServer(http.HandlerFunc(handlers.ServeHTTP(metadataService.GetV2Handler())))
	defer testServer.Close()

	// make a request to the testing server
	res, err := http.Get(fmt.Sprintf("%s/v2/metadata/", testServer.URL))
	assert.NoError(t, err, "Unexpected error making HTTP Request")
	response, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	assert.NoError(t, err, "Unexpected error reading HTTP response")

	actualMetadata := &v2.TaskResponse{}
	err = json.Unmarshal(response, actualMetadata)
	assert.NoError(t, err, "Unexpected error unmarshalling response")

	assert.ElementsMatch(t, expectedMetadata.Containers, actualMetadata.Containers, "Expected container responses to match")
	assert.Equal(t, expectedMetadata.TaskTags, actualMetadata.TaskTags, "Expected Task Tags to match")
	assert.Equal(t, expectedMetadata.ContainerInstanceTags, actualMetadata.ContainerInstanceTags, "Expected Container Instance Tags to match")
	assert.Equal(t, expectedMetadata.Cluster, actualMetadata.Cluster, "Expected Cluster to match")
	assert.Equal(t, expectedMetadata.Family, actualMetadata.Family, "Expected Family to match")
	assert.Equal(t, expectedMetadata.Revision, actualMetadata.Revision, "Expected Revision to match")
	assert.Equal(t, expectedMetadata.DesiredStatus, actualMetadata.DesiredStatus, "Expected DesiredStatus to match")
	assert.Equal(t, expectedMetadata.KnownStatus, actualMetadata.KnownStatus, "Expected KnownStatus to match")

}
