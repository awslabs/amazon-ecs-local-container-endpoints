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

package main

import (
	"fmt"
	"net/http"

	"github.com/awslabs/amazon-ecs-local-container-endpoints/local-container-endpoints/config"
	"github.com/awslabs/amazon-ecs-local-container-endpoints/local-container-endpoints/handlers"
	"github.com/awslabs/amazon-ecs-local-container-endpoints/local-container-endpoints/utils"
	"github.com/awslabs/amazon-ecs-local-container-endpoints/local-container-endpoints/version"
	"github.com/sirupsen/logrus"
)

func main() {
	logrus.Info(version.String())
	logrus.Info("Running...")
	credentialsService, err := handlers.NewCredentialService()
	if err != nil {
		logrus.Fatal("Failed to create Credentials Service: ", err)
	}

	metadataService, err := handlers.NewMetadataService()
	if err != nil {
		logrus.Fatal("Failed to create Metadata Service: ", err)
	}

	http.HandleFunc("/role/", handlers.ServeHTTP(credentialsService.GetRoleHandler()))
	http.HandleFunc("/creds", handlers.ServeHTTP(credentialsService.GetTemporaryCredentialHandler()))

	http.HandleFunc("/v2/", handlers.ServeHTTP(metadataService.GetV2Handler()))
	http.HandleFunc("/ecs-local-metadata-v3/", handlers.ServeHTTP(metadataService.GetV3Handler()))

	port := utils.GetValue(config.DefaultPort, config.PortVar)
	err = http.ListenAndServe(fmt.Sprintf(":%s", port), nil)
	if err != nil {
		logrus.Fatal("HTTP Server exited with error: ", err)
	}
}
