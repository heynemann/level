// level game server
// https://github.com/heynemann/level
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Bernardo Heynemann <heynemann@gmail.com>

// Source file: https://github.com/nats-io/nats/blob/master/test/test.go
// Copyright 2015 Apcera Inc. All rights reserved.

package pubsub_test

import (
	"fmt"

	"github.com/onsi/gomega/types"
)

func mapEqual(m1, m2 map[string]interface{}) bool {
	if len(m1) != len(m2) {
		return false
	}

	for k, v := range m1 {
		if val, ok := m2[k]; ok {
			vStr := fmt.Sprintf("%v", v)
			valStr := fmt.Sprintf("%v", val)
			if vStr != valStr {
				return false
			}
		} else {
			return false
		}
	}

	return true
}

func MapEqual(expected map[string]interface{}) types.GomegaMatcher {
	return &mapEqualMatcher{
		expected: expected,
	}
}

type mapEqualMatcher struct {
	expected map[string]interface{}
}

func (matcher *mapEqualMatcher) Match(actual interface{}) (success bool, err error) {
	ok := mapEqual(matcher.expected, actual.(map[string]interface{}))
	if !ok {
		return false, fmt.Errorf("%v is not the same as %v", matcher.expected, actual)
	}

	return true, nil
}

func (matcher *mapEqualMatcher) FailureMessage(actual interface{}) (message string) {
	return fmt.Sprintf("Expected\n\t%#v\nto be the same as \n\t%#v", actual, matcher.expected)
}

func (matcher *mapEqualMatcher) NegatedFailureMessage(actual interface{}) (message string) {
	return fmt.Sprintf("Expected\n\t%#v\nnot to be the same as \n\t%#v", actual, matcher.expected)
}
