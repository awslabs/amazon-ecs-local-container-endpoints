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
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/awslabs/amazon-ecs-local-container-endpoints/local-container-endpoints/clients/iam/mock_iamiface"
	"github.com/awslabs/amazon-ecs-local-container-endpoints/local-container-endpoints/clients/sts/mock_stsiface"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

const (
	roleName             = "clyde_task_role"
	roleARN              = "arn:aws:iam::111111111111111:role/clyde_task_role"
	secretKey            = "SKID"
	accessKey            = "AKID"
	sessionToken         = "token"
	expirationTimeString = "2009-11-10T23:00:00Z"
)

func TestGetRoleCredentials(t *testing.T) {
	iamMock, stsMock := setupMocks(t)

	credsService := newCredentialServiceInTest(iamMock, stsMock)

	expiration, _ := time.Parse(credentialExpirationTimeFormat, expirationTimeString)

	gomock.InOrder(
		iamMock.EXPECT().GetRole(gomock.Any()).Do(func(x interface{}) {
			input := x.(*iam.GetRoleInput)
			assert.Equal(t, roleName, aws.StringValue(input.RoleName), "Expected role name to match")
		}).Return(&iam.GetRoleOutput{
			Role: &iam.Role{
				Arn: aws.String(roleARN),
			},
		}, nil),
		stsMock.EXPECT().AssumeRole(gomock.Any()).Do(func(x interface{}) {
			input := x.(*sts.AssumeRoleInput)
			assert.Equal(t, roleARN, aws.StringValue(input.RoleArn), "Expected role ARN to match")
		}).Return(&sts.AssumeRoleOutput{
			Credentials: &sts.Credentials{
				AccessKeyId:     aws.String(accessKey),
				SecretAccessKey: aws.String(secretKey),
				SessionToken:    aws.String(sessionToken),
				Expiration:      &expiration,
			},
		}, nil),
	)

	response, err := credsService.getRoleCredentials(fmt.Sprintf("/role/%s", roleName))
	assert.NoError(t, err, "Unexpected error calling getRoleCredentials")
	assert.Equal(t, response.AccessKeyId, accessKey, "Expected access key to match")
	assert.Equal(t, response.SecretAccessKey, secretKey, "Expected secret key to match")
	assert.Equal(t, response.Token, sessionToken, "Expected session token to match")
	assert.Equal(t, response.Expiration, expirationTimeString, "Expected expiration to match")
	assert.Equal(t, response.RoleArn, roleARN, "Expected role ARN to match")

}

func TestGetRoleCredentialsInvalidURL(t *testing.T) {
	iamMock, stsMock := setupMocks(t)

	credsService := newCredentialServiceInTest(iamMock, stsMock)

	_, err := credsService.getRoleCredentials("/role/*")
	assert.Error(t, err, "Expected error calling getRoleCredentials")

}

func TestGetRoleCredentialsGetRoleError(t *testing.T) {
	iamMock, stsMock := setupMocks(t)

	credsService := newCredentialServiceInTest(iamMock, stsMock)

	gomock.InOrder(
		iamMock.EXPECT().GetRole(gomock.Any()).Do(func(x interface{}) {
			input := x.(*iam.GetRoleInput)
			assert.Equal(t, roleName, aws.StringValue(input.RoleName), "Expected role name to match")
		}).Return(nil, fmt.Errorf("Some API Error")),
	)

	_, err := credsService.getRoleCredentials(fmt.Sprintf("/role/%s", roleName))
	assert.Error(t, err, "Expected error calling getRoleCredentials")

}

func TestGetRoleCredentialsSTSError(t *testing.T) {
	iamMock, stsMock := setupMocks(t)

	credsService := newCredentialServiceInTest(iamMock, stsMock)

	gomock.InOrder(
		iamMock.EXPECT().GetRole(gomock.Any()).Do(func(x interface{}) {
			input := x.(*iam.GetRoleInput)
			assert.Equal(t, roleName, aws.StringValue(input.RoleName), "Expected role name to match")
		}).Return(&iam.GetRoleOutput{
			Role: &iam.Role{
				Arn: aws.String(roleARN),
			},
		}, nil),
		stsMock.EXPECT().AssumeRole(gomock.Any()).Do(func(x interface{}) {
			input := x.(*sts.AssumeRoleInput)
			assert.Equal(t, roleARN, aws.StringValue(input.RoleArn), "Expected role ARN to match")
		}).Return(nil, fmt.Errorf("Some API Error")),
	)

	_, err := credsService.getRoleCredentials(fmt.Sprintf("/role/%s", roleName))
	assert.Error(t, err, "Expected error calling getRoleCredentials")

}

func TestGetTemporaryCredentials(t *testing.T) {
	iamMock, stsMock := setupMocks(t)

	credsService := newCredentialServiceInTest(iamMock, stsMock)

	expiration, _ := time.Parse(credentialExpirationTimeFormat, expirationTimeString)

	gomock.InOrder(
		stsMock.EXPECT().GetSessionToken(gomock.Any()).Return(&sts.GetSessionTokenOutput{
			Credentials: &sts.Credentials{
				AccessKeyId:     aws.String(accessKey),
				SecretAccessKey: aws.String(secretKey),
				SessionToken:    aws.String(sessionToken),
				Expiration:      &expiration,
			},
		}, nil),
	)

	response, err := credsService.getTemporaryCredentials()
	assert.NoError(t, err, "Unexpected error calling getRoleCredentials")
	assert.Equal(t, response.AccessKeyId, accessKey, "Expected access key to match")
	assert.Equal(t, response.SecretAccessKey, secretKey, "Expected secret key to match")
	assert.Equal(t, response.Token, sessionToken, "Expected session token to match")
	assert.Equal(t, response.Expiration, expirationTimeString, "Expected expiration to match")

}

func TestGetTemporaryCredentialsErrorCase(t *testing.T) {
	iamMock, stsMock := setupMocks(t)

	credsService := newCredentialServiceInTest(iamMock, stsMock)

	gomock.InOrder(
		stsMock.EXPECT().GetSessionToken(gomock.Any()).Return(nil, fmt.Errorf("Some API Error")),
	)

	_, err := credsService.getTemporaryCredentials()
	assert.Error(t, err, "Expected error calling getRoleCredentials")

}

type CustomProvider struct {
	expiration time.Time
	creds      credentials.Value
}

func (m *CustomProvider) Retrieve() (credentials.Value, error) {
	return m.creds, nil
}
func (m *CustomProvider) IsExpired() bool {
	return false
}
func (m *CustomProvider) ExpiresAt() time.Time {
	return m.expiration
}

func TestGetTemporaryCredentialsExistingTempCreds(t *testing.T) {
	expiration, _ := time.Parse(credentialExpirationTimeFormat, expirationTimeString)

	provider := &CustomProvider{
		expiration: expiration,
		creds: credentials.Value{
			AccessKeyID:     accessKey,
			SecretAccessKey: secretKey,
			SessionToken:    sessionToken,
		},
	}

	creds := credentials.NewCredentials(provider)
	svcConfig := aws.NewConfig().WithCredentials(creds)
	sess, err := session.NewSession(svcConfig)
	assert.NoError(t, err, "Unexpected error creating new session")

	credsService := &CredentialService{
		currentSession: sess,
	}

	response, err := credsService.getTemporaryCredentials()
	assert.NoError(t, err, "Unexpected error calling getRoleCredentials")
	assert.Equal(t, response.AccessKeyId, accessKey, "Expected access key to match")
	assert.Equal(t, response.SecretAccessKey, secretKey, "Expected secret key to match")
	assert.Equal(t, response.Token, sessionToken, "Expected session token to match")
	assert.Equal(t, response.Expiration, expirationTimeString, "Expected expiration to match")

}

func setupMocks(t *testing.T) (*mock_iamiface.MockIAMAPI, *mock_stsiface.MockSTSAPI) {
	ctrl := gomock.NewController(t)
	iamMock := mock_iamiface.NewMockIAMAPI(ctrl)
	stsMock := mock_stsiface.NewMockSTSAPI(ctrl)
	return iamMock, stsMock
}

func newCredentialServiceInTest(iamMock *mock_iamiface.MockIAMAPI, stsMock *mock_stsiface.MockSTSAPI) *CredentialService {
	return &CredentialService{
		stsClient:      stsMock,
		iamClient:      iamMock,
		currentSession: nil,
	}
}
