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

// Package Docker includes a wrapper of the Docker Go SDK Client
package docker

import (
	"context"
	"os"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

const (
	// v1.27 is the oldest API version
	// which has all the latest changes to the APIs we use.
	minDockerAPIVersion = "1.27"
)

// Client is a wrapper for Docker SDK Client
type Client interface {
	ContainerList(context.Context) ([]types.Container, error)
}

type dockerClient struct {
	sdkClient *client.Client
}

// NewDockerClient creates a new wrapper of the Docker Go Client
func NewDockerClient() (Client, error) {
	// Using NewEnvClient allows customers to configure Docker via env vars
	// However, if DOCKER_API_VERSION is not set, the SDK can pick a version
	// which is too new for the local Docker.
	if os.Getenv("DOCKER_API_VERSION") == "" {
		os.Setenv("DOCKER_API_VERSION", minDockerAPIVersion)
	}
	sdkClient, err := client.NewEnvClient()
	if err != nil {
		return nil, err
	}
	return &dockerClient{
		sdkClient: sdkClient,
	}, nil
}

func (c *dockerClient) ContainerList(ctx context.Context) ([]types.Container, error) {
	return c.sdkClient.ContainerList(ctx, types.ContainerListOptions{})
}
