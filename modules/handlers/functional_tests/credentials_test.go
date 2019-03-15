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

package functional_tests

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/awslabs/amazon-ecs-local-container-endpoints/modules/clients/iam/mock_iamiface"
	"github.com/awslabs/amazon-ecs-local-container-endpoints/modules/clients/sts/mock_stsiface"
	"github.com/awslabs/amazon-ecs-local-container-endpoints/modules/handlers"
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

	expiration, _ := time.Parse(time.RFC3339, expirationTimeString)

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

	ts := httptest.NewServer(http.HandlerFunc(handlers.ServeHTTP(credsService.GetRoleHandler())))
	defer ts.Close()

	res, err := http.Get(fmt.Sprintf("%s/role/%s", ts.URL, roleName))
	assert.NoError(t, err, "Unexpected error making HTTP Request")
	response, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	assert.NoError(t, err, "Unexpected error reading HTTP response")

	creds := &handlers.CredentialResponse{}
	err = json.Unmarshal(response, creds)
	assert.NoError(t, err, "Unexpected error unmarshalling response")
	assert.Equal(t, creds.AccessKeyId, accessKey, "Expected access key to match")
	assert.Equal(t, creds.SecretAccessKey, secretKey, "Expected secret key to match")
	assert.Equal(t, creds.Token, sessionToken, "Expected session token to match")
	assert.Equal(t, creds.Expiration, expirationTimeString, "Expected expiration to match")
	assert.Equal(t, creds.RoleArn, roleARN, "Expected role ARN to match")
}

func TestGetTemporaryCredentials(t *testing.T) {
	iamMock, stsMock := setupMocks(t)

	credsService := newCredentialServiceInTest(iamMock, stsMock)

	expiration, _ := time.Parse(time.RFC3339, expirationTimeString)

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

	ts := httptest.NewServer(http.HandlerFunc(handlers.ServeHTTP(credsService.GetTemporaryCredentialHandler())))
	defer ts.Close()

	res, err := http.Get(fmt.Sprintf("%s/role/%s", ts.URL, roleName))
	assert.NoError(t, err, "Unexpected error making HTTP Request")
	response, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	assert.NoError(t, err, "Unexpected error reading HTTP response")

	creds := &handlers.CredentialResponse{}
	err = json.Unmarshal(response, creds)

	assert.NoError(t, err, "Unexpected error calling getRoleCredentials")
	assert.Equal(t, creds.AccessKeyId, accessKey, "Expected access key to match")
	assert.Equal(t, creds.SecretAccessKey, secretKey, "Expected secret key to match")
	assert.Equal(t, creds.Token, sessionToken, "Expected session token to match")
	assert.Equal(t, creds.Expiration, expirationTimeString, "Expected expiration to match")

}

func setupMocks(t *testing.T) (*mock_iamiface.MockIAMAPI, *mock_stsiface.MockSTSAPI) {
	ctrl := gomock.NewController(t)
	iamMock := mock_iamiface.NewMockIAMAPI(ctrl)
	stsMock := mock_stsiface.NewMockSTSAPI(ctrl)
	return iamMock, stsMock
}

func newCredentialServiceInTest(iamMock *mock_iamiface.MockIAMAPI, stsMock *mock_stsiface.MockSTSAPI) *handlers.CredentialService {
	return handlers.NewCredentialServiceWithClients(iamMock, stsMock, nil)
}
