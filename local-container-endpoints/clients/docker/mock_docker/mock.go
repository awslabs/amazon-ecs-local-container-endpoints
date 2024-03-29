// Copyright 2019 Amazon.com, Inc. or its affiliates. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License"). You may
// not use this file except in compliance with the License. A copy of the
// License is located at
//
//     http://aws.amazon.com/apache2.0/
//
// or in the "license" file accompanying this file. This file is distributed
// on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either
// express or implied. See the License for the specific language governing
// permissions and limitations under the License.

// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/awslabs/amazon-ecs-local-container-endpoints/local-container-endpoints/clients/docker (interfaces: Client)

// Package mock_docker is a generated GoMock package.
package mock_docker

import (
	context "context"
	reflect "reflect"

	types "github.com/docker/docker/api/types"
	gomock "github.com/golang/mock/gomock"
)

// MockClient is a mock of Client interface.
type MockClient struct {
	ctrl     *gomock.Controller
	recorder *MockClientMockRecorder
}

// MockClientMockRecorder is the mock recorder for MockClient.
type MockClientMockRecorder struct {
	mock *MockClient
}

// NewMockClient creates a new mock instance.
func NewMockClient(ctrl *gomock.Controller) *MockClient {
	mock := &MockClient{ctrl: ctrl}
	mock.recorder = &MockClientMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockClient) EXPECT() *MockClientMockRecorder {
	return m.recorder
}

// ContainerList mocks base method.
func (m *MockClient) ContainerList(arg0 context.Context) ([]types.Container, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ContainerList", arg0)
	ret0, _ := ret[0].([]types.Container)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ContainerList indicates an expected call of ContainerList.
func (mr *MockClientMockRecorder) ContainerList(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ContainerList", reflect.TypeOf((*MockClient)(nil).ContainerList), arg0)
}

// ContainerStats mocks base method.
func (m *MockClient) ContainerStats(arg0 context.Context, arg1 string) (*types.Stats, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ContainerStats", arg0, arg1)
	ret0, _ := ret[0].(*types.Stats)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ContainerStats indicates an expected call of ContainerStats.
func (mr *MockClientMockRecorder) ContainerStats(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ContainerStats", reflect.TypeOf((*MockClient)(nil).ContainerStats), arg0, arg1)
}
