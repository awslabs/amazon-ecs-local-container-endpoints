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

	"github.com/awslabs/amazon-ecs-local-container-endpoints/local-container-endpoints/testingutils"
	"github.com/docker/docker/api/types"
	"github.com/stretchr/testify/assert"
)

const (
	endpointsShortID = "56771b9219b5"
	endpointsLongID  = "56771b9219b58c8b6a286830667b62475e79753db34a0b82a98efafb20718c0f9"
	shortID1         = "e18ab3d25b38"
	longID1          = "e18ab3d25b38c8b6a287831767b62475a79853dc38a0b92a98efabb20718c0d90"
	longID2          = "457129ed3bd03f1fc70125c3be7bcbee760d5edf092e32155a5c6a730cd32020"
	longID3          = "0756a2371cad1976b07954490660f07d240a6a6f52d17594ed691799915695f7"
	containerName1   = "container1-puddles"
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

func TestFindContainerWithIdentifierID(t *testing.T) {
	endpointsContainer := testingutils.BaseDockerContainer("endpoints", endpointsLongID).Get()
	container1 := testingutils.BaseDockerContainer("caller", longID1).Get()
	container2 := testingutils.BaseDockerContainer("pudding", longID2).Get()
	container3 := testingutils.BaseDockerContainer("dumpling", longID3).Get()

	containers := []types.Container{
		endpointsContainer,
		container1,
		container2,
		container3,
	}

	var testCases = []struct {
		identifier        string
		expectedContainer *types.Container
	}{
		{
			identifier:        shortID1,
			expectedContainer: &container1,
		},
		{
			identifier:        longID2,
			expectedContainer: &container2,
		},
		{
			identifier:        longID3,
			expectedContainer: &container3,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.identifier, func(t *testing.T) {
			actual, err := findContainer(containers, testCase.identifier, "")
			assert.NoError(t, err, "Unexpected error from findContainer")
			assert.Equal(t, testCase.expectedContainer, actual, "Expected findContainer to find the correct container")
		})
	}

}

func TestFindContainerWithIdentifierName(t *testing.T) {
	endpointsContainer := testingutils.BaseDockerContainer("endpoints", endpointsLongID).Get()
	container1 := testingutils.BaseDockerContainer(containerName1, longID1).Get()
	container2 := testingutils.BaseDockerContainer(containerName2, longID2).Get()
	container3 := testingutils.BaseDockerContainer(containerName3, longID3).Get()

	containers := []types.Container{
		container2,
		container1,
		endpointsContainer,
		container3,
	}

	var testCases = []struct {
		identifier        string
		expectedContainer *types.Container
	}{
		{
			identifier:        containerName1,
			expectedContainer: &container1,
		},
		{
			identifier:        containerName2,
			expectedContainer: &container2,
		},
		{
			identifier:        containerName3,
			expectedContainer: &container3,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.identifier, func(t *testing.T) {
			actual, err := findContainer(containers, testCase.identifier, "")
			assert.NoError(t, err, "Unexpected error from findContainer")
			assert.Equal(t, testCase.expectedContainer, actual, "Expected findContainer to find the correct container")
		})
	}

}

func TestFindContainerWithCallerIP(t *testing.T) {
	endpointsContainer := testingutils.BaseDockerContainer("endpoints", endpointsLongID).WithNetwork("bridge", ipAddress).Get()
	container1 := testingutils.BaseDockerContainer(containerName1, longID1).WithNetwork("bridge", ipAddress1).Get()
	container2 := testingutils.BaseDockerContainer(containerName2, longID2).WithNetwork("bridge", ipAddress2).Get()
	container3 := testingutils.BaseDockerContainer(containerName3, longID3).WithNetwork("bridge", ipAddress3).Get()

	containers := []types.Container{
		container3,
		container1,
		container2,
		endpointsContainer,
	}

	var testCases = []struct {
		callerIP          string
		expectedContainer *types.Container
	}{
		{
			callerIP:          ipAddress1,
			expectedContainer: &container1,
		},
		{
			callerIP:          ipAddress2,
			expectedContainer: &container2,
		},
		{
			callerIP:          ipAddress3,
			expectedContainer: &container3,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.callerIP, func(t *testing.T) {
			actual, err := findContainer(containers, "", testCase.callerIP)
			assert.NoError(t, err, "Unexpected error from findContainer")
			assert.Equal(t, testCase.expectedContainer, actual, "Expected findContainer to find the correct container")
		})
	}

}

func TestFindContainerWithCallerIPAndNetworks(t *testing.T) {
	os.Setenv("HOSTNAME", endpointsShortID)
	defer os.Unsetenv("HOSTNAME")

	endpointsContainer := testingutils.BaseDockerContainer("endpoints", endpointsLongID).WithNetwork(network1, ipAddress).Get()
	container1 := testingutils.BaseDockerContainer(containerName1, longID1).WithNetwork(network2, ipAddress1).Get()
	container2 := testingutils.BaseDockerContainer(containerName2, longID2).WithNetwork(network1, ipAddress2).Get()
	container3 := testingutils.BaseDockerContainer(containerName3, longID3).WithNetwork(network1, ipAddress1).Get()

	containers := []types.Container{
		container3,
		container1,
		container2,
		endpointsContainer,
	}

	actual, err := findContainer(containers, "", ipAddress1)
	assert.NoError(t, err, "Unexpected error from findContainer")
	assert.Equal(t, &container3, actual, "Expected findContainer to find the correct container")

}

func TestFindContainerWithCallerIPAndNetworksFailure(t *testing.T) {
	endpointsContainer := testingutils.BaseDockerContainer("endpoints", endpointsLongID).WithNetwork("bridge", ipAddress).Get()
	container1 := testingutils.BaseDockerContainer(containerName1, longID1).WithNetwork(network2, ipAddress1).Get()
	container2 := testingutils.BaseDockerContainer(containerName2, longID2).WithNetwork(network1, ipAddress2).Get()
	container3 := testingutils.BaseDockerContainer(containerName3, longID3).WithNetwork(network1, ipAddress1).Get()

	containers := []types.Container{
		container3,
		container1,
		container2,
		endpointsContainer,
	}

	_, err := findContainer(containers, "", ipAddress1)
	// No container matches
	assert.Error(t, err, "Expected error from findContainer")

}

// An unlikely scenario in which all of the checks (identifier, IP, and networks) must work correctly in order for the right container to be returned
func TestFindContainerWithAllChecks(t *testing.T) {
	endpointsContainer := testingutils.BaseDockerContainer("endpoints", endpointsLongID).WithNetwork(network1, ipAddress).WithNetwork(network2, ipAddress).Get()
	container1 := testingutils.BaseDockerContainer(containerName1, longID1).WithNetwork(network2, ipAddress1).Get()
	container2 := testingutils.BaseDockerContainer(containerName2, longID2).WithNetwork(network1, ipAddress2).Get()
	container3 := testingutils.BaseDockerContainer(containerName3, longID3).WithNetwork(network1, ipAddress1).Get()

	containers := []types.Container{
		container3,
		container1,
		container2,
		endpointsContainer,
	}

	// container 1 & 2 are matched by the identifier "pud", and container 1 & 3 have ipAddress1 in a valid network
	actual, err := findContainer(containers, "pud", ipAddress1)
	assert.NoError(t, err, "Unexpected error from findContainer")
	assert.Equal(t, &container1, actual, "Expected findContainer to find the correct container")

	// error cases to prove that both identifier and ip were needed:
	_, err = findContainer(containers, "pud", "")
	assert.Error(t, err, "Expected error from findContainer")

	_, err = findContainer(containers, "", ipAddress1)
	assert.Error(t, err, "Expected error from findContainer")

}

func TestFindContainerFailureMoreThanOneMatches(t *testing.T) {
	endpointsContainer := testingutils.BaseDockerContainer("endpoints", endpointsLongID).WithNetwork(network1, ipAddress).WithNetwork(network2, ipAddress).Get()
	container1 := testingutils.BaseDockerContainer(containerName1, longID1).WithNetwork(network2, ipAddress1).Get()
	container2 := testingutils.BaseDockerContainer(containerName2, longID2).WithNetwork(network1, ipAddress2).Get()
	container3 := testingutils.BaseDockerContainer(containerName3, longID3).WithNetwork(network1, ipAddress1).Get()

	containers := []types.Container{
		container3,
		container1,
		container2,
		endpointsContainer,
	}

	// all the containers have 'container' in their name, and endpoints has two networks so the IPAddress doesn't identify the container either
	_, err := findContainer(containers, "container", ipAddress1)
	// No container matches
	assert.Error(t, err, "Expected error from findContainer")

}

func TestFindContainerWithIdentifierFailure(t *testing.T) {
	endpointsContainer := testingutils.BaseDockerContainer("endpoints", endpointsLongID).WithNetwork("bridge", ipAddress).Get()
	container1 := testingutils.BaseDockerContainer(containerName1, longID1).Get()
	container2 := testingutils.BaseDockerContainer(containerName2, longID2).Get()
	container3 := testingutils.BaseDockerContainer(containerName3, longID3).Get()

	containers := []types.Container{
		container3,
		container1,
		container2,
		endpointsContainer,
	}

	_, err := findContainer(containers, badName, "")
	// No container matches
	assert.Error(t, err, "Expected error from findContainer")

}

func TestGetTaskContainers(t *testing.T) {
	endpointsContainer := testingutils.BaseDockerContainer("endpoints", endpointsLongID).WithNetwork(network1, ipAddress).WithComposeProject(projectName).Get()
	container1 := testingutils.BaseDockerContainer(containerName1, longID1).WithNetwork(network2, ipAddress1).WithComposeProject(projectName2).Get()
	container2 := testingutils.BaseDockerContainer(containerName2, longID2).WithNetwork(network1, ipAddress2).WithComposeProject(projectName).Get()
	container3 := testingutils.BaseDockerContainer(containerName3, longID3).WithNetwork(network1, ipAddress1).WithComposeProject(projectName).Get()

	containers := []types.Container{
		container3,
		container1,
		container2,
		endpointsContainer,
	}

	expected := []types.Container{
		container3,
		container2,
		endpointsContainer,
	}

	result := getTaskContainers(containers, "", ipAddress1)

	assert.ElementsMatch(t, expected, result, "Expected containers returned by getTaskContainers to be from the correct compose project")

}

// Pass in nil for any value which is allowed to be nil, to verify that the code can never panic
func TestGetTaskContainersNilTest(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Error(r)
		}
	}()
	endpointsContainer := testingutils.BaseDockerContainer("endpoints", endpointsLongID).WithNetwork(network1, ipAddress).WithNetwork(network2, ipAddress).WithNetwork("bridge", ipAddress).WithComposeProject(projectName2).Get()
	container1 := testingutils.BaseDockerContainer(containerName3, longID1).WithNetwork(network2, ipAddress1).WithComposeProject(projectName2).Get()
	container2 := testingutils.BaseDockerContainer(containerName3, longID2).WithNetwork(network1, ipAddress2).WithComposeProject(projectName2).Get()
	container3 := testingutils.BaseDockerContainer(containerName3, longID3).WithNetwork(network1, ipAddress1).WithComposeProject(projectName).Get()

	endpointsContainer.NetworkSettings.Networks["bridge"] = nil
	container1.NetworkSettings.Networks[network1] = nil
	container2.Names = nil
	container3.Labels = nil

	containers := []types.Container{
		container3,
		container1,
		container2,
		endpointsContainer,
	}

	getTaskContainers(containers, containerName3, ipAddress1)
}

func TestGetTaskContainersOneContainerReturned(t *testing.T) {
	// technically
	endpointsContainer := testingutils.BaseDockerContainer("endpoints", endpointsLongID).WithNetwork(network1, ipAddress).WithComposeProject(projectName2).Get()
	container1 := testingutils.BaseDockerContainer(containerName1, longID1).WithNetwork(network2, ipAddress1).WithComposeProject(projectName2).Get()
	container2 := testingutils.BaseDockerContainer(containerName2, longID2).WithNetwork(network1, ipAddress2).WithComposeProject(projectName2).Get()
	container3 := testingutils.BaseDockerContainer(containerName3, longID3).WithNetwork(network1, ipAddress1).WithComposeProject(projectName).Get()

	containers := []types.Container{
		container3,
		container1,
		container2,
		endpointsContainer,
	}

	expected := []types.Container{
		container3,
	}

	result := getTaskContainers(containers, containerName3, ipAddress1)

	assert.ElementsMatch(t, expected, result, "Expected containers returned by getTaskContainers to be from the correct compose project")

}

// TODO: re-enable test once metadata with Tags field is added
// func TestNewMetadataServiceWithTags(t *testing.T) {
// 	os.Setenv(config.ContainerInstanceTagsVar, "mitchell=webb,thats=numberwang")
// 	os.Setenv(config.TaskTagsVar, "hello=goodbye,get=back,come=together")
// 	defer os.Clearenv()
//
// 	expectedCITags := map[string]string{
// 		"mitchell": "webb",
// 		"thats":    "numberwang",
// 	}
// 	expectedTaskTags := map[string]string{
// 		"hello": "goodbye",
// 		"get":   "back",
// 		"come":  "together",
// 	}
//
// 	service, err := NewMetadataService()
// 	assert.NoError(t, err, "Unexpected error calling NewMetadataService")
// 	assert.Equal(t, expectedCITags, service.containerInstanceTags, "Expected container instance tags to match")
// 	assert.Equal(t, expectedTaskTags, service.taskTags, "Expected task tags to match")
// }
