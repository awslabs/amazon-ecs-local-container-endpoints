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
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/iam/iamiface"
	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/aws/aws-sdk-go/service/sts/stsiface"
	"github.com/awslabs/amazon-ecs-local-container-endpoints/local-container-endpoints/clients/useragent"
	"github.com/awslabs/amazon-ecs-local-container-endpoints/local-container-endpoints/config"
	"github.com/awslabs/amazon-ecs-local-container-endpoints/local-container-endpoints/utils"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

const (
	temporaryCredentialsDurationInS = 3600
	roleSessionNameLength           = 64
)

const (
	// CredentialExpirationTimeFormat is the time stamp format used in the Local Credentials Service HTTP response
	CredentialExpirationTimeFormat = time.RFC3339
)

// CredentialService vends credentials to containers
type CredentialService struct {
	iamClient      iamiface.IAMAPI
	stsClient      stsiface.STSAPI
	currentSession *session.Session
}

// NewCredentialService returns a struct that handles credentials requests
func NewCredentialService() (*CredentialService, error) {
	iamCustomEndpoint := utils.GetValue("", config.IAMCustomEndpointVar)
	if iamCustomEndpoint != "" {
		logrus.Infof("Using custom IAM endpoint %s", iamCustomEndpoint)
	}

	stsCustomEndpoint := utils.GetValue("", config.STSCustomEndpointVar)
	if stsCustomEndpoint != "" {
		logrus.Infof("Using custom STS endpoint %s", stsCustomEndpoint)
	}

	defaultResolver := endpoints.DefaultResolver()
	customResolverFn := func(service, region string, optFns ...func(*endpoints.Options)) (endpoints.ResolvedEndpoint, error) {
		if service == endpoints.IamServiceID && iamCustomEndpoint != "" {
			return endpoints.ResolvedEndpoint{
				URL: iamCustomEndpoint,
			}, nil
		} else if service == endpoints.StsServiceID && stsCustomEndpoint != "" {
			return endpoints.ResolvedEndpoint{
				URL: stsCustomEndpoint,
			}, nil
		}
		return defaultResolver.EndpointFor(service, region, optFns...)
	}

	sess, err := session.NewSessionWithOptions(session.Options{
		Config: aws.Config{
			EndpointResolver:              endpoints.ResolverFunc(customResolverFn),
			CredentialsChainVerboseErrors: aws.Bool(true),
		},
		SharedConfigState: session.SharedConfigEnable,
	})
	if err != nil {
		return nil, err
	}

	iamClient := iam.New(sess)
	iamClient.Handlers.Build.PushBackNamed(useragent.CustomUserAgentHandler())
	stsClient := sts.New(sess)
	stsClient.Handlers.Build.PushBackNamed(useragent.CustomUserAgentHandler())
	return NewCredentialServiceWithClients(iamClient, stsClient, sess), nil
}

// NewCredentialServiceWithClients returns a struct that handles credentials requests with the given clients
func NewCredentialServiceWithClients(iamClient iamiface.IAMAPI, stsClient stsiface.STSAPI, currentSession *session.Session) *CredentialService {
	return &CredentialService{
		iamClient:      iamClient,
		stsClient:      stsClient,
		currentSession: currentSession,
	}
}

// SetupRoutes sets up the credentials paths in mux
func (service *CredentialService) SetupRoutes(router *mux.Router) {
	router.HandleFunc(config.RoleCredentialsPath, ServeHTTP(service.getRoleHandler()))
	router.HandleFunc(config.RoleCredentialsPathWithSlash, ServeHTTP(service.getRoleHandler()))

	router.HandleFunc(config.RoleArnCredentialsPath, ServeHTTP(service.getRoleArnHandler()))
	router.HandleFunc(config.RoleArnCredentialsPathWithSlash, ServeHTTP(service.getRoleArnHandler()))

	router.HandleFunc(config.TempCredentialsPath, ServeHTTP(service.getTemporaryCredentialHandler()))
	router.HandleFunc(config.TempCredentialsPathWithSlash, ServeHTTP(service.getTemporaryCredentialHandler()))
}

// GetRoleHandler returns the Task IAM Role handler
func (service *CredentialService) getRoleHandler() func(w http.ResponseWriter, r *http.Request) error {
	return func(w http.ResponseWriter, r *http.Request) error {
		logrus.Debug("Received role credentials request")

		vars := mux.Vars(r)
		roleName := vars["role"]
		if roleName == "" {
			return HTTPError{
				Code: http.StatusBadRequest,
				Err:  fmt.Errorf("Invalid URL path %s; expected '/role/<IAM Role Name>'", r.URL.Path),
			}
		}

		response, err := service.getRoleCredentials(roleName)
		if err != nil {
			return err
		}

		writeJSONResponse(w, response)
		return nil
	}
}

// GetRoleArnHandler returns the Task IAM Role handler for complete role ARNs
func (service *CredentialService) getRoleArnHandler() func(w http.ResponseWriter, r *http.Request) error {
	return func(w http.ResponseWriter, r *http.Request) error {
		logrus.Debug("Received role credentials request using ARN")

		vars := mux.Vars(r)
		roleName := vars["roleName"]
		roleArn := fmt.Sprintf("%s/%s", vars["roleArn"], roleName)
		if roleArn == "" {
			return HTTPError{
				Code: http.StatusBadRequest,
				Err:  fmt.Errorf("Invalid URL path %s; expected '/role-arn/<IAM Role ARN>", r.URL.Path),
			}
		}

		response, err := service.getRoleCredentialsFromArn(roleArn, roleName)
		if err != nil {
			return err
		}

		writeJSONResponse(w, response)
		return nil
	}
}

func (service *CredentialService) getRoleCredentials(roleName string) (*CredentialResponse, error) {
	logrus.Debugf("Requesting credentials for %s", roleName)

	output, err := service.iamClient.GetRole(&iam.GetRoleInput{
		RoleName: aws.String(roleName),
	})
	if err != nil {
		return nil, err
	}

	return service.getRoleCredentialsFromArn(aws.StringValue(output.Role.Arn), roleName)
}

func (service *CredentialService) getRoleCredentialsFromArn(roleArn, roleName string) (*CredentialResponse, error) {
	logrus.Debugf("Requesting credentials for role with ARN %s", roleArn)

	creds, err := service.stsClient.AssumeRole(&sts.AssumeRoleInput{
		RoleArn:         aws.String(roleArn),
		DurationSeconds: aws.Int64(temporaryCredentialsDurationInS),
		RoleSessionName: aws.String(utils.Truncate(fmt.Sprintf("ecs-local-%s", roleName), roleSessionNameLength)),
	})

	if err != nil {
		return nil, err
	}

	return &CredentialResponse{
		AccessKeyID:     aws.StringValue(creds.Credentials.AccessKeyId),
		SecretAccessKey: aws.StringValue(creds.Credentials.SecretAccessKey),
		RoleArn:         roleArn,
		Token:           aws.StringValue(creds.Credentials.SessionToken),
		Expiration:      creds.Credentials.Expiration.Format(CredentialExpirationTimeFormat),
	}, nil
}

// GetTemporaryCredentialHandler returns a handler which vends temporary credentials for the local IAM identity
func (service *CredentialService) getTemporaryCredentialHandler() func(w http.ResponseWriter, r *http.Request) error {
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
		credVal, err := service.currentSession.Config.Credentials.Get()
		if err != nil {
			return nil, errors.Wrap(err, "Current session is based on temporary credentials, but they were not retrieved.")
		}

		logrus.Debug("Current session contains temporary credentials")
		response := CredentialResponse{
			AccessKeyID:     credVal.AccessKeyID,
			SecretAccessKey: credVal.SecretAccessKey,
			Token:           credVal.SessionToken,
		}

		expiration, err := service.currentSession.Config.Credentials.ExpiresAt()

		// It is valid for a credential provider to not return an expiration;
		// however, we need to have an expiration if a token is present to
		// satsify various client SDKs. In this case, we return an expiration
		// timestamp a fixed point in the future.
		// https://github.com/awslabs/amazon-ecs-local-container-endpoints/issues/26
		if err != nil && len(response.Token) > 0 {
			expiration, err = getSharedTokenExpiration()
		}

		if err == nil {
			response.Expiration = expiration.Format(CredentialExpirationTimeFormat)
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
		AccessKeyID:     aws.StringValue(creds.Credentials.AccessKeyId),
		SecretAccessKey: aws.StringValue(creds.Credentials.SecretAccessKey),
		Token:           aws.StringValue(creds.Credentials.SessionToken),
		Expiration:      creds.Credentials.Expiration.Format(CredentialExpirationTimeFormat),
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

// Return an expiration date a set point in the future. error is currently
// always nil (we gracefully fail back to the 12.5 minute default), but we
// reserve it for future use in case there are valid reasons to error out.
func getSharedTokenExpiration() (time.Time, error) {
	durationStr := utils.GetValue(fmt.Sprintf("%ds", config.DefaultSharedTokenExpiration), config.SharedTokenExpirationVar)
	duration, err := time.ParseDuration(durationStr)

	if err != nil {
		// If they didn't provide a unit, try to parse this as seconds.
		durationSeconds, err := strconv.ParseInt(durationStr, 0, 64)
		if err != nil {
			logrus.Warnf(
				"Could not parse SHARED_TOKEN_EXPIRATION value, defaulting to %d seconds: %s",
				config.DefaultSharedTokenExpiration, durationStr)
			durationSeconds = config.DefaultSharedTokenExpiration
		}

		duration = time.Duration(durationSeconds) * time.Second
	}

	// Make sure the duration is always in the future.
	if duration <= 0 {
		logrus.Warnf(
			"SHARED_TOKEN_EXPIRATION value must be positive, forcing to %d seconds: %s",
			config.DefaultSharedTokenExpiration, durationStr)
		duration = config.DefaultSharedTokenExpiration * time.Second
	}

	return time.Now().UTC().Add(duration), nil
}
