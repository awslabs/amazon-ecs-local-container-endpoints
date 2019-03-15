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

// These tests create a session and make a request to AWS to verify that the SDK can uptake credentials
// from the local credentials service.
package credentials

import (
	"os"
	"testing"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/stretchr/testify/assert"
)

func TestCredentials_TemporaryCredentials(t *testing.T) {
	os.Setenv("AWS_CONTAINER_CREDENTIALS_RELATIVE_URI", "/creds")
	defer os.Setenv("AWS_CONTAINER_CREDENTIALS_RELATIVE_URI", "")
	sess, err := session.NewSession()
	assert.NoError(t, err, "Unexpected error creating an SDK session")

	s3Client := s3.New(sess)

	output, err := s3Client.ListBuckets(&s3.ListBucketsInput{})
	assert.NoError(t, err, "Unexpected error calling list buckets")
	assert.NotNil(t, output, "Expected list bucket response to be non-nil")
}

func TestCredentials_RoleCredentials(t *testing.T) {
	os.Setenv("AWS_CONTAINER_CREDENTIALS_RELATIVE_URI", "/role/ecs-local-endpoints-integ-role")
	defer os.Setenv("AWS_CONTAINER_CREDENTIALS_RELATIVE_URI", "")
	sess, err := session.NewSession()
	assert.NoError(t, err, "Unexpected error creating an SDK session")

	s3Client := s3.New(sess)

	output, err := s3Client.ListBuckets(&s3.ListBucketsInput{})
	assert.NoError(t, err, "Unexpected error calling list buckets")
	assert.NotNil(t, output, "Expected list bucket response to be non-nil")
}

// Just to verify that the SDK is taking creds from the Local Credentials Service
// Set AWS_CONTAINER_CREDENTIALS_RELATIVE_URI to something that will throw an error
func TestCredentials_Error(t *testing.T) {
	os.Setenv("AWS_CONTAINER_CREDENTIALS_RELATIVE_URI", "/cats")
	defer os.Setenv("AWS_CONTAINER_CREDENTIALS_RELATIVE_URI", "")
	sess, err := session.NewSession()
	assert.NoError(t, err, "Unexpected error creating an SDK session")

	s3Client := s3.New(sess)

	// Failure will occur on the actual API call
	_, err = s3Client.ListBuckets(&s3.ListBucketsInput{})
	assert.Error(t, err, "Expected error calling list buckets")
}
