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
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/awslabs/amazon-ecs-local-container-endpoints/local-container-endpoints/config"
	"github.com/awslabs/amazon-ecs-local-container-endpoints/local-container-endpoints/handlers"
	"github.com/awslabs/amazon-ecs-local-container-endpoints/local-container-endpoints/utils"
	"github.com/awslabs/amazon-ecs-local-container-endpoints/local-container-endpoints/version"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

func main() {
	logrus.Info(version.String())
	logrus.Info("Running...")
	credentialsService, err := handlers.NewCredentialService()
	if err != nil {
		logrus.Fatal("Failed to create Credentials Service: ", err)
	}

	contMetadata := getBaseMetadata(config.ContainerMetadataPathVar)
	taskMetadata := getBaseMetadata(config.TaskMetadataPathVar)

	metadataService, err := handlers.NewMetadataService(taskMetadata, contMetadata)
	if err != nil {
		logrus.Fatal("Failed to create Metadata Service: ", err)
	}

	port := utils.GetValue(config.DefaultPort, config.PortVar)

	router := mux.NewRouter()
	metadataService.SetupV2Routes(router)
	metadataService.SetupV3Routes(router)
	credentialsService.SetupRoutes(router)

	server := http.Server{
		Addr:    fmt.Sprintf(":%s", port),
		Handler: router,
	}
	err = server.ListenAndServe()
	if err != nil {
		logrus.Fatal("HTTP Server exited with error: ", err)
	}
}

func getBaseMetadata(pathVar string) map[string]interface{} {
	path := os.Getenv(pathVar)
	if path == "" {
		return nil
	}

	metadataFile, err := os.Open(path)
	if err != nil {
		logrus.Error("Failed to read user defined metadata file: ", err)
		return nil
	}

	bits, err := ioutil.ReadAll(metadataFile)
	if err != nil {
		logrus.Error("Failed to read user defined metadata file: ", err)
		return nil
	}

	var metadata map[string]interface{}
	err = json.Unmarshal(bits, &metadata)
	if err != nil {
		logrus.Error("Failed to read user defined metadata file: ", err)
		return nil
	}

	return metadata
}
