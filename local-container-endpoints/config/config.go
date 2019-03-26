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

// Package config contains environment variables and default values for Local Endpoints
package config

// Environment Variables
const (
	// PortEnvVar defines the port that metadata and credentials listen at
	PortVar = "ECS_LOCAL_METADATA_PORT"

	// Metadata related
	ClusterARNVar            = "CLUSTER_ARN"
	TaskARNVar               = "TASK_ARN"
	TDFamilyVar              = "TASK_DEFINITION_FAMILY"
	TDRevisionVar            = "TASK_DEFINITION_REVISION"
	ContainerInstanceTagsVar = "CONTAINER_INSTANCE_TAGS"
	TaskTagsVar              = "TASK_TAGS_VAR"
)

// Defaults
const (
	// DefaultPort is the default port the server listens at
	DefaultPort = "80"

	// Metadata related
	DefaultContainerType = "NORMAL"
	DefaultClusterName   = "ecs-local-cluster"
	DefaultTaskARN       = "arn:aws:ecs:us-west-2:111111111111:task/ecs-local-cluster/37e873f6-37b4-42a7-af47-eac7275c6152"
	DefaultTDFamily      = "esc-local-task-definition"
	DefaultTDRevision    = "1"
)

// Settings
const (
	HTTPTimeoutDuration = "5s"
)

// URL Paths

// Credentials
const (
	// RoleCredentialsPath is the path for obtaining credentials from a role
	RoleCredentialsPath = "/role/{role}"
	// RoleCredentialsPathWithSlash adds a trailing slash
	RoleCredentialsPathWithSlash = RoleCredentialsPath + "/"

	// TempCredentialsPath is the path for obtaining temp creds from sts:GetSessionsToken
	TempCredentialsPath = "/creds"
	// TempCredentialsPathWithSlash adds a trailing slash
	TempCredentialsPathWithSlash = TempCredentialsPath + "/"
)

// V3
const (
	// V3ContainerMetadataPath is the path for V3 container metadata
	V3ContainerMetadataPath = "/v3"
	// V3ContainerMetadataPathWithSlash adds a trailing slash
	V3ContainerMetadataPathWithSlash = V3ContainerMetadataPath + "/"
	// V3ContainerMetadataPathWithIdentifier is the V3 container metadata path with an identifer specified
	V3ContainerMetadataPathWithIdentifier = "/v3/containers/{identifier}"
	// V3ContainerMetadataPathWithIdentifierAndSlash adds a trailing slash
	V3ContainerMetadataPathWithIdentifierAndSlash = V3ContainerMetadataPathWithIdentifier + "/"

	// V3ContainerStatsPath is the path for V3 container stats
	V3ContainerStatsPath = "/v3/stats"
	// V3ContainerStatsPathWithSlash adds a trailing slash
	V3ContainerStatsPathWithSlash = V3ContainerStatsPath + "/"
	// V3ContainerStatsPathWithIdentifier is the V3 container stats path with an identifier
	V3ContainerStatsPathWithIdentifier = "/v3/containers/{identifier}/stats"
	// V3ContainerStatsPathWithIdentifierAndSlash adds a trailing slash
	V3ContainerStatsPathWithIdentifierAndSlash = V3ContainerStatsPathWithIdentifier + "/"

	// V3TaskMetadataPath is the path for V3 task metadata
	V3TaskMetadataPath = "/v3/task"
	// V3TaskMetadataPathWithSlash adds a trailing slash
	V3TaskMetadataPathWithSlash = V3TaskMetadataPath + "/"
	// V3TaskMetadataPathWithIdentifier is the v3 task metadata path with an identifier
	V3TaskMetadataPathWithIdentifier = "/v3/containers/{identifier}/task"
	// V3TaskMetadataPathWithIdentifierWithSlash adds a trailing slash
	V3TaskMetadataPathWithIdentifierWithSlash = V3TaskMetadataPathWithIdentifier + "/"

	// V3TaskStatsPath is the path for V3 task stats
	V3TaskStatsPath = "/v3/task/stats"
	// V3TaskStatsPathWithSlash adds a trailing slash
	V3TaskStatsPathWithSlash = V3TaskStatsPath + "/"
	// V3TaskStatsPathWithIdentifier is the v3 task stats path with an identifier
	V3TaskStatsPathWithIdentifier = "/v3/containers/{identifier}/task/stats"
	// V3TaskStatsPathWithIdentifierAndSlash adds a trailing slash
	V3TaskStatsPathWithIdentifierAndSlash = V3TaskStatsPathWithIdentifier + "/"
)

// V2
const (
	// V2TaskMetadataPath is the V2 Task Metadata path
	V2TaskMetadataPath = "/v2/metadata"
	// V2TaskMetadataPathWithSlash adds a trailing slash
	V2TaskMetadataPathWithSlash = V2TaskMetadataPath + "/"

	// V2ContainerMetadataPath is the V2 Container Metadata path
	V2ContainerMetadataPath = "/v2/metadata/{identifier}"
	// V2ContainerMetadataPathWithSlash adds a trailing slash
	V2ContainerMetadataPathWithSlash = V2ContainerMetadataPath + "/"

	// V2TaskStatsPath is the V2 Task Stats Path
	V2TaskStatsPath = "/v2/stats"
	// V2TaskStatsPathWithSlash adds a trailing slash
	V2TaskStatsPathWithSlash = V2TaskStatsPath + "/"

	// V2ContainerStatsPath is the V2 container stats path
	V2ContainerStatsPath = "/v2/stats/{identifier}"
	// V2ContainerStatsPathWithSlash adds a trailing slash
	V2ContainerStatsPathWithSlash = V2ContainerStatsPath + "/"
)
