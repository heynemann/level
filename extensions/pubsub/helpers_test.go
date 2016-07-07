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
)

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
// Creating client connections
////////////////////////////////////////////////////////////////////////////////

// NewDefaultConnection
func NewDefaultConnection() *nats.Conn {
	return NewConnection(nats.DefaultPort)
}

// NewConnection forms connection on a given port.
func NewConnection(port int) *nats.Conn {
	url := fmt.Sprintf("nats://localhost:%d", port)
	nc, err := nats.Connect(url)
	if err != nil {
		return nil
	}
	return nc
}

// NewEConn
func NewEConn() *nats.EncodedConn {
	ec, err := nats.NewEncodedConn(NewDefaultConnection(), nats.DEFAULT_ENCODER)
	if err != nil {
		return nil
	}
	return ec
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
