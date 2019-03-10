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

// Package handlers containers the HTTP Handlers for Local Metadata and Credentials
package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/sirupsen/logrus"
)

// Error wraps built-in error and adds a status code
type Error interface {
	error
	Status() int
}

// HttpError represents an error with a HTTP status code.
type HttpError struct {
	Code int
	Err  error
}

// Error() satisfies the error interface.
func (herr HttpError) Error() string {
	return herr.Err.Error()
}

// Status retutns the HTTP status code.
func (herr HttpError) Status() int {
	return herr.Code
}

// ServeHTTP wraps an HTTP Handler
func ServeHTTP(handler func(w http.ResponseWriter, r *http.Request) error) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		err := handler(w, r)
		if err != nil {
			switch e := err.(type) {
			case Error:
				// Return the specific error code and error message
				logrus.Errorf("HTTP %d - %s", e.Status(), err)
				http.Error(w, e.Error(), e.Status())
			default:
				// default to HTTP 500 for all other errors
				logrus.Errorf("HTTP 500 - %s", err)
				// Internal Server Error: <actual error message>
				http.Error(w, fmt.Sprintf("%s: %s", http.StatusText(http.StatusInternalServerError), err.Error()),
					http.StatusInternalServerError)
			}
		}
	}
}

func writeJSONResponse(w http.ResponseWriter, response interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
