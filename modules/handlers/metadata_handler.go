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
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/awslabs/amazon-ecs-local-container-endpoints/modules/clients/docker"
	"github.com/awslabs/amazon-ecs-local-container-endpoints/modules/config"
	"github.com/awslabs/amazon-ecs-local-container-endpoints/modules/metadata"
	"github.com/awslabs/amazon-ecs-local-container-endpoints/modules/utils"
	"github.com/docker/docker/api/types"
	"github.com/sirupsen/logrus"
)

const (
	composeProjectNameLabel = "com.docker.compose.project"
)

// MetadataService vends docker metadata to containers
type MetadataService struct {
	dockerClient          docker.Client
	containerInstanceTags map[string]string
	taskTags              map[string]string
}

// NewMetadataService returns a struct that handles metadata requests
func NewMetadataService() (*MetadataService, error) {
	dockerClient, err := docker.NewDockerClient()
	if err != nil {
		return nil, err
	}
	return NewMetadataServiceWithClient(dockerClient)
}

// NewMetadataServiceWithClient returns a struct that handles metadata requests using the given Docker Client
func NewMetadataServiceWithClient(dockerClient docker.Client) (*MetadataService, error) {
	metadata := &MetadataService{
		dockerClient: dockerClient,
	}

	if ciTagVal := os.Getenv(config.ContainerInstanceTagsVar); ciTagVal != "" {
		tags, err := utils.GetTagsMap(ciTagVal)
		if err != nil {
			return nil, err
		}
		metadata.containerInstanceTags = tags
	}

	if taskTagVal := os.Getenv(config.TaskTagsVar); taskTagVal != "" {
		tags, err := utils.GetTagsMap(taskTagVal)
		if err != nil {
			return nil, err
		}
		metadata.taskTags = tags
	}

	return metadata, nil
}

// GetV3Handler returns the task metadata V3 Handler
func (service *MetadataService) GetV3Handler() func(w http.ResponseWriter, r *http.Request) error {
	return func(w http.ResponseWriter, r *http.Request) error {
		callerIP, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			// Failed to get the callerIP
			callerIP = ""
		}

		// URL Path format = /ecs-local-metadata-v3/<container identifier>/
		//                   /ecs-local-metadata-v3/<container identifier>/stats
		//                   /ecs-local-metadata-v3/<container identifier>/task/
		//                   /ecs-local-metadata-v3/<container identifier>/task/stats
		if re := regexp.MustCompile("/ecs-local-metadata-v3/([a-zA-Z0-9_-]*)/task/?"); re.MatchString(r.URL.Path) {
			// return val should be [full URL Path, container identifier]
			urlParts := re.FindStringSubmatch(r.URL.Path)

			if len(urlParts) < 2 {
				return HttpError{
					Code: http.StatusBadRequest,
					Err:  fmt.Errorf("Invalid URL path %s; expected '/ecs-local-metadata-v3/<container identifier>/task/'", r.URL.Path),
				}
			}
			identifier := urlParts[1]
			return service.taskMetadataResponse(w, identifier, callerIP)
		}

		return HttpError{
			Code: http.StatusBadRequest,
			Err:  fmt.Errorf("invalid URL path: %s", r.URL.Path),
		}
	}
}

// GetV2Handler returns the task metadata V2 Handler
func (service *MetadataService) GetV2Handler() func(w http.ResponseWriter, r *http.Request) error {
	return func(w http.ResponseWriter, r *http.Request) error {
		callerIP, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			// Failed to get the callerIP
			callerIP = ""
		}

		if re := regexp.MustCompile("/v2/metadata/?"); re.MatchString(r.URL.Path) {
			return service.taskMetadataResponse(w, "", callerIP)
		}

		return HttpError{
			Code: http.StatusBadRequest,
			Err:  fmt.Errorf("invalid URL path: %s", r.URL.Path),
		}
	}
}

func (service *MetadataService) taskMetadataResponse(w http.ResponseWriter, identifier string, callerIP string) error {
	timeout, _ := time.ParseDuration(config.HTTPTimeoutDuration)
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	containers, err := service.dockerClient.ContainerList(ctx)
	if err != nil {
		return err
	}
	taskContainers := getTaskContainers(containers, identifier, callerIP)

	response := metadata.GetTaskMetadata(taskContainers, service.containerInstanceTags, service.taskTags)

	writeJSONResponse(w, response)
	return nil
}

// A Local 'Task' is defined as all containers in the same Docker Compose Project as the caller container
// OR all containers running on this machine if the user is not using Compose
func getTaskContainers(allContainers []types.Container, identifier string, callerIP string) []types.Container {
	callerContainer, err := findContainer(allContainers, identifier, callerIP)
	if err != nil {
		logrus.Warn(err)
		logrus.Info("Will use all containers to represent one 'local task'")
		return allContainers
	}

	projectName := callerContainer.Labels[composeProjectNameLabel]

	if projectName == "" {
		logrus.Info("Will use all containers to represent one 'local task': The container which made the request is not in a Docker Compose Project")
		return allContainers
	}

	return filterByComposeProject(allContainers, projectName)
}

func filterByComposeProject(dockerContainers []types.Container, projectName string) []types.Container {
	var filteredContainers []types.Container

	for _, container := range dockerContainers {
		if container.Labels[composeProjectNameLabel] == projectName {
			filteredContainers = append(filteredContainers, container)
		}
	}

	if len(filteredContainers) > 0 {
		return filteredContainers
	}

	return dockerContainers
}

// Algorithm:
// 1. Given a list of all running containers
// 2. Filter the list by the <container identifier> if it was present in the request URI. If this leaves only one container, then we have found our match.
// 	a. First we check if the identifier was is a prefix for the container ID (i.e. it is the container short ID or the full ID), and then we check if it was a subset of the container name
// 3. Filter the remaining results in the list by the request IP. If this leaves only one container, then we have found our match.
// 4. Filter the remaining results by the docker networks that the endpoint container is in. A container can only call the endpoints if it is in the same docker network as the endpoints container.
// 	a. Determine which Docker Networks the Endpoints container is in by determining which container it is (We can do this using $HOSTNAME, which will be our container short ID) and then use the output of Docker API's ContainerList (https://godoc.org/github.com/docker/docker/client#Client.ContainerList) to find its networks.
// 	b. Filter the remaining containers by selecting those containers which have the callerIP in one of the endpoints container's networks.
// 5. If no container is found, or more than one container matches, we return an error.
func findContainer(dockerContainers []types.Container, identifier string, callerIP string) (*types.Container, error) {
	var filteredList []types.Container = dockerContainers

	if identifier != "" {
		filteredList = filterContainersByIdentifier(dockerContainers, identifier)
		if len(filteredList) == 1 { // we found the container
			return &filteredList[0], nil
		}
	}

	if callerIP != "" {
		filteredList = filterContainersByRequestIP(filteredList, callerIP)
		if len(filteredList) == 1 { // we found the container
			return &filteredList[0], nil
		}
	}

	filteredList = filterContainersByMyNetworks(filteredList, dockerContainers, callerIP)
	if len(filteredList) == 1 { // we found the container
		return &filteredList[0], nil
	}

	return nil, fmt.Errorf("Failed to find the container which the request came from. Narrowed down search to %d containers", len(filteredList))
}

func filterContainersByIdentifier(dockerContainers []types.Container, identifier string) []types.Container {
	var filteredList []types.Container
	for _, container := range dockerContainers {
		if strings.HasPrefix(container.ID, identifier) {
			filteredList = append(filteredList, container)
			continue
		}

		for _, name := range container.Names {
			if strings.Contains(name, identifier) {
				filteredList = append(filteredList, container)
			}
		}
	}
	if len(filteredList) > 0 {
		return filteredList
	}
	return dockerContainers

}

func filterContainersByRequestIP(dockerContainers []types.Container, callerIP string) []types.Container {
	var filteredList []types.Container
	for _, container := range dockerContainers {
		if container.NetworkSettings == nil {
			continue
		}
		for _, settings := range container.NetworkSettings.Networks {
			if settings != nil && settings.IPAddress == callerIP {
				filteredList = append(filteredList, container)
			}
		}

	}

	if len(filteredList) > 0 {
		return filteredList
	}
	return dockerContainers
}

// filter the list by the networks which the endpoints container is in
func filterContainersByMyNetworks(filteredContainerList []types.Container, allContainers []types.Container, callerIP string) []types.Container {
	// find endpoints containers
	var endpointContainer *types.Container
	shortID := os.Getenv("HOSTNAME")
	for _, container := range allContainers {
		if strings.HasPrefix(container.ID, shortID) {
			endpointContainer = &container
		}
	}

	if endpointContainer == nil || endpointContainer.NetworkSettings == nil {
		logrus.Warn("Failed to find endpoints container among running containers")
		// Return the list we were given, since we can't filter it any further
		return filteredContainerList
	}

	var finalList []types.Container

	// containers can only make request to the endpoint container from within one of its networks
	var networksToSearch []string
	for network, settings := range endpointContainer.NetworkSettings.Networks {
		networksToSearch = append(networksToSearch, network)
		networksToSearch = append(networksToSearch, settings.Aliases...)
	}

	for _, container := range filteredContainerList {
		if container.NetworkSettings == nil {
			continue
		}
		for network, settings := range container.NetworkSettings.Networks {
			if settings != nil && networkMatches(network, settings.Aliases, networksToSearch) && settings.IPAddress == callerIP {
				// This container is in one of the right networks and has the caller IP in that network
				finalList = append(finalList, container)
			}
		}
	}

	return finalList
}

// Returns true if the networkName of any alias is in the list networksToSearch
func networkMatches(networkName string, aliases []string, networksToSearch []string) bool {
	for _, check := range networksToSearch {
		if networkName == check {
			return true
		}
		for _, alias := range aliases {
			if alias == check {
				return true
			}
		}
	}

	return false
}
