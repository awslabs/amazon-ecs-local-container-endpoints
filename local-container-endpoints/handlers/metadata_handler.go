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
	"fmt"
	"net"
	"net/http"
	"os"

	"github.com/awslabs/amazon-ecs-local-container-endpoints/local-container-endpoints/clients/docker"
	"github.com/awslabs/amazon-ecs-local-container-endpoints/local-container-endpoints/config"
	"github.com/awslabs/amazon-ecs-local-container-endpoints/local-container-endpoints/utils"
	"github.com/gorilla/mux"
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

// SetupV2Routes sets up the V2 Metadata routes
func (service *MetadataService) SetupV2Routes(router *mux.Router) {
	router.HandleFunc(config.V2TaskMetadataPath, ServeHTTP(service.getMetadataHandler(requestTypeTaskMetadata)))
	router.HandleFunc(config.V2TaskMetadataPathWithSlash, ServeHTTP(service.getMetadataHandler(requestTypeTaskMetadata)))

	router.HandleFunc(config.V2TaskStatsPath, ServeHTTP(service.getMetadataHandler(requestTypeTaskStats)))
	router.HandleFunc(config.V2TaskStatsPathWithSlash, ServeHTTP(service.getMetadataHandler(requestTypeTaskStats)))

	router.HandleFunc(config.V2ContainerMetadataPath, ServeHTTP(service.getMetadataHandler(requestTypeContainerMetadata)))
	router.HandleFunc(config.V2ContainerMetadataPathWithSlash, ServeHTTP(service.getMetadataHandler(requestTypeContainerMetadata)))

	router.HandleFunc(config.V2ContainerStatsPath, ServeHTTP(service.getMetadataHandler(requestTypeContainerStats)))
	router.HandleFunc(config.V2ContainerStatsPathWithSlash, ServeHTTP(service.getMetadataHandler(requestTypeContainerStats)))
}

// SetupV3Routes sets up the V3 Metadata routes
func (service *MetadataService) SetupV3Routes(router *mux.Router) {
	router.HandleFunc(config.V3ContainerMetadataPath, ServeHTTP(service.getMetadataHandler(requestTypeContainerMetadata)))
	router.HandleFunc(config.V3ContainerMetadataPathWithSlash, ServeHTTP(service.getMetadataHandler(requestTypeContainerMetadata)))
	router.HandleFunc(config.V3ContainerMetadataPathWithIdentifier, ServeHTTP(service.getMetadataHandler(requestTypeContainerMetadata)))
	router.HandleFunc(config.V3ContainerMetadataPathWithIdentifierAndSlash, ServeHTTP(service.getMetadataHandler(requestTypeContainerMetadata)))

	router.HandleFunc(config.V3ContainerStatsPath, ServeHTTP(service.getMetadataHandler(requestTypeContainerStats)))
	router.HandleFunc(config.V3ContainerStatsPathWithSlash, ServeHTTP(service.getMetadataHandler(requestTypeContainerStats)))
	router.HandleFunc(config.V3ContainerStatsPathWithIdentifier, ServeHTTP(service.getMetadataHandler(requestTypeContainerStats)))
	router.HandleFunc(config.V3ContainerStatsPathWithIdentifierAndSlash, ServeHTTP(service.getMetadataHandler(requestTypeContainerStats)))

	router.HandleFunc(config.V3TaskMetadataPath, ServeHTTP(service.getMetadataHandler(requestTypeTaskMetadata)))
	router.HandleFunc(config.V3TaskMetadataPathWithSlash, ServeHTTP(service.getMetadataHandler(requestTypeTaskMetadata)))
	router.HandleFunc(config.V3TaskMetadataPathWithIdentifier, ServeHTTP(service.getMetadataHandler(requestTypeTaskMetadata)))
	router.HandleFunc(config.V3TaskMetadataPathWithIdentifierWithSlash, ServeHTTP(service.getMetadataHandler(requestTypeTaskMetadata)))

	router.HandleFunc(config.V3TaskStatsPath, ServeHTTP(service.getMetadataHandler(requestTypeTaskStats)))
	router.HandleFunc(config.V3TaskStatsPathWithSlash, ServeHTTP(service.getMetadataHandler(requestTypeTaskStats)))
	router.HandleFunc(config.V3TaskStatsPathWithIdentifier, ServeHTTP(service.getMetadataHandler(requestTypeTaskStats)))
	router.HandleFunc(config.V3TaskStatsPathWithIdentifierAndSlash, ServeHTTP(service.getMetadataHandler(requestTypeTaskStats)))
}

// getMetadataHandler returns a metadata handler given a requestType
func (service *MetadataService) getMetadataHandler(requestType int) func(w http.ResponseWriter, r *http.Request) error {
	return func(w http.ResponseWriter, r *http.Request) error {
		callerIP, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			// Failed to get the callerIP
			callerIP = ""
		}
		vars := mux.Vars(r)
		identifier := vars["identifier"]
		return service.handleRequest(requestType, w, identifier, callerIP)
	}
}

func (service *MetadataService) handleRequest(requestType int, w http.ResponseWriter, identifier string, callerIP string) error {
	switch requestType {
	case requestTypeTaskMetadata:
		return service.taskMetadataResponse(w, identifier, callerIP)
	case requestTypeTaskStats:
		return service.taskStatsResponse(w, identifier, callerIP)
	case requestTypeContainerStats:
		return service.containerStatsResponse(w, identifier, callerIP)
	case requestTypeContainerMetadata:
		return service.containerMetadataResponse(w, identifier, callerIP)
	}

	// This should never run, but explicitly returning an error here helps make it easy to find bugs
	return fmt.Errorf("There's a bug in this code: Invalid request type %d", requestType)
}
