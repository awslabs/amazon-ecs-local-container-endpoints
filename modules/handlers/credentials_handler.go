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
	"net/http"
	"regexp"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/iam/iamiface"
	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/aws/aws-sdk-go/service/sts/stsiface"
	"github.com/awslabs/amazon-ecs-local-container-endpoints/modules/utils"
	"github.com/sirupsen/logrus"
)

const (
	temporaryCredentialsDurationInS = 3600
	roleSessionNameLength           = 64
	credentialExpirationTimeFormat  = time.RFC3339
)

// CredentialService vends credentials to containers
type CredentialService struct {
	iamClient      iamiface.IAMAPI
	stsClient      stsiface.STSAPI
	currentSession *session.Session
}

// NewCredentialService returns a struct that handles credentials requests
func NewCredentialService() (*CredentialService, error) {
	sess, err := session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	})
	if err != nil {
		return nil, err
	}
	return NewCredentialServiceWithClients(iam.New(sess), sts.New(sess), sess), nil
}

// NewCredentialServiceWithClients returns a struct that handles credentials requests with the given clients
func NewCredentialServiceWithClients(iamClient iamiface.IAMAPI, stsClient stsiface.STSAPI, currentSession *session.Session) *CredentialService {
	return &CredentialService{
		iamClient:      iamClient,
		stsClient:      stsClient,
		currentSession: currentSession,
	}
}

// GetRoleHandler returns the Task IAM Role handler
func (service *CredentialService) GetRoleHandler() func(w http.ResponseWriter, r *http.Request) error {
	return func(w http.ResponseWriter, r *http.Request) error {
		logrus.Debug("Received role credentials request")
		response, err := service.getRoleCredentials(r.URL.Path)
		if err != nil {
			return err
		}

		writeJSONResponse(w, response)
		return nil
	}
}

func (service *CredentialService) getRoleCredentials(urlPath string) (*CredentialResponse, error) {
	// URL Path format = /role/<role name>
	regExpr := regexp.MustCompile(`/role/([\w+=,.@-]+)`)
	urlParts := regExpr.FindStringSubmatch(urlPath)

	if len(urlParts) < 2 {
		return nil, HttpError{
			Code: http.StatusBadRequest,
			Err:  fmt.Errorf("Invalid URL path %s; expected '/role/<IAM Role Name>'", urlPath),
		}
	}

	roleName := urlParts[1]
	logrus.Debugf("Requesting credentials for %s", roleName)

	output, err := service.iamClient.GetRole(&iam.GetRoleInput{
		RoleName: aws.String(roleName),
	})
	if err != nil {
		return nil, err
	}

	creds, err := service.stsClient.AssumeRole(&sts.AssumeRoleInput{
		RoleArn:         output.Role.Arn,
		DurationSeconds: aws.Int64(temporaryCredentialsDurationInS),
		RoleSessionName: aws.String(utils.Truncate(fmt.Sprintf("ecs-local-%s", roleName), roleSessionNameLength)),
	})

	if err != nil {
		return nil, err
	}

	return &CredentialResponse{
		AccessKeyId:     aws.StringValue(creds.Credentials.AccessKeyId),
		SecretAccessKey: aws.StringValue(creds.Credentials.SecretAccessKey),
		RoleArn:         aws.StringValue(output.Role.Arn),
		Token:           aws.StringValue(creds.Credentials.SessionToken),
		Expiration:      creds.Credentials.Expiration.Format(credentialExpirationTimeFormat),
	}, nil
}

// GetTemporaryCredentialHandler returns a handler which vends temporary credentials for the local IAM identity
func (service *CredentialService) GetTemporaryCredentialHandler() func(w http.ResponseWriter, r *http.Request) error {
	return func(w http.ResponseWriter, r *http.Request) error {
		logrus.Debug("Received temporary local credentials request")

		response, err := service.getTemporaryCredentials()
		if err != nil {
			return err
		}

		writeJSONResponse(w, response)
		return nil
	}
}

func (service *CredentialService) getTemporaryCredentials() (*CredentialResponse, error) {
	// check if the current session already was built on temp creds
	// because temp creds do not have the power to call GetSessionToken
	if service.isCurrentSessionTemporary() {
		credVal, _ := service.currentSession.Config.Credentials.Get()

		logrus.Debug("Current session contains temporary credentials")
		response := CredentialResponse{
			AccessKeyId:     credVal.AccessKeyID,
			SecretAccessKey: credVal.SecretAccessKey,
			Token:           credVal.SessionToken,
		}
		expiration, err := service.currentSession.Config.Credentials.ExpiresAt()
		// It is valid for a credential provider to not return an expiration
		// TODO: Check if expiration is optional from the POV of the SDKs
		if err == nil {
			response.Expiration = expiration.Format(credentialExpirationTimeFormat)
		}
		return &response, nil
	}

	// current session is not temp creds, so we can call GetSessionToken
	creds, err := service.stsClient.GetSessionToken(&sts.GetSessionTokenInput{
		DurationSeconds: aws.Int64(temporaryCredentialsDurationInS),
	})

	if err != nil {
		return nil, err
	}

	response := CredentialResponse{
		AccessKeyId:     aws.StringValue(creds.Credentials.AccessKeyId),
		SecretAccessKey: aws.StringValue(creds.Credentials.SecretAccessKey),
		Token:           aws.StringValue(creds.Credentials.SessionToken),
		Expiration:      creds.Credentials.Expiration.Format(credentialExpirationTimeFormat),
	}

	return &response, nil
}

func (service *CredentialService) isCurrentSessionTemporary() bool {
	if service.currentSession != nil && service.currentSession.Config != nil && service.currentSession.Config.Credentials != nil {
		credVal, err := service.currentSession.Config.Credentials.Get()

		if err == nil && credVal.SessionToken != "" { // current session is already temp creds
			return true
		}
	}
	return false
}
