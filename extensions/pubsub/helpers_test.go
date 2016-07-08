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
	"errors"
	"fmt"
	"time"

	gnatsdServer "github.com/nats-io/gnatsd/server"
	gnatsdTest "github.com/nats-io/gnatsd/test"
	"github.com/nats-io/nats"
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

// Dumb wait program to sync on callbacks, etc... Will timeout
func Wait(ch chan bool) error {
	return WaitTime(ch, 5*time.Second)
}

// Wait for a chan with a timeout.
func WaitTime(ch chan bool, timeout time.Duration) error {
	select {
	case <-ch:
		return nil
	case <-time.After(timeout):
	}
	return errors.New("timeout")
}

////////////////////////////////////////////////////////////////////////////////
// Running gnatsd server in separate Go routines
////////////////////////////////////////////////////////////////////////////////

// RunDefaultServer will run a server on the default port.
func RunDefaultServer() *gnatsdServer.Server {
	return RunServerOnPort(nats.DefaultPort)
}

// RunServerOnPort will run a server on the given port.
func RunServerOnPort(port int) *gnatsdServer.Server {
	opts := gnatsdTest.DefaultTestOptions
	opts.Port = port
	return RunServerWithOptions(opts)
}

// RunServerWithOptions will run a server with the given options.
func RunServerWithOptions(opts gnatsdServer.Options) *gnatsdServer.Server {
	return gnatsdTest.RunServer(&opts)
}

// RunServerWithConfig will run a server with the given configuration file.
func RunServerWithConfig(configFile string) (*gnatsdServer.Server, *gnatsdServer.Options) {
	return gnatsdTest.RunServerWithConfig(configFile)
}
