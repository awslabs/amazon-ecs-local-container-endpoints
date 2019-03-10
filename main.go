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
	"os"

	"github.com/awslabs/amazon-ecs-local-container-endpoints/modules/config"
	"github.com/awslabs/amazon-ecs-local-container-endpoints/modules/handlers"
	"github.com/awslabs/amazon-ecs-local-container-endpoints/modules/version"
	"github.com/sirupsen/logrus"
)

func main() {
	logrus.Info(version.String())
	logrus.Info("Running...")
	credentialsService, err := handlers.NewCredentialService()
	if err != nil {
		logrus.Fatal("Failed to create Credentials Server: ", err)
	}

	http.HandleFunc("/role/", handlers.ServeHTTP(credentialsService.GetRoleHandler()))
	http.HandleFunc("/creds", handlers.ServeHTTP(credentialsService.GetTemporaryCredentialHandler()))

	port := config.DefaultPort
	if os.Getenv(config.PortEnvVar) != "" {
		port = os.Getenv(config.PortEnvVar)
	}
	err = http.ListenAndServe(fmt.Sprintf(":%s", port), nil)
	if err != nil {
		logrus.Fatal("HTTP Server exited with error: ", err)
	}
}
