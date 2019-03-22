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

// Package utils is a grab bag of things that don't belong anywhere else
package utils

import (
	"fmt"
	"os"
	"strings"
)

// Truncate truncates a string
func Truncate(s string, length int) string {
	if len(s) > length {
		return s[0:length]
	}

	return s
}

// GetTagsMap parses tags in the format key1=value1,key2=value2
func GetTagsMap(value string) (map[string]string, error) {
	tags := make(map[string]string)
	keyValPairs := strings.Split(value, ",")
	for _, pair := range keyValPairs {
		split := strings.Split(pair, "=")
		if len(split) != 2 {
			return nil, fmt.Errorf("Tag input not formatted correctly: %s", pair)
		}
		tags[split[0]] = split[1]
	}
	return tags, nil
}

// GetValue Returns the value of the envVar, or the default
func GetValue(defaultVal, envVar string) string {
	if val := os.Getenv(envVar); val != "" {
		return val
	}

	return defaultVal
}
