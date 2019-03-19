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
	PortEnvVar = "ECS_LOCAL_METADATA_PORT"
)

// Defaults
const (
	// DefaultPort is the default port the server listens at
	DefaultPort = "80"
)
