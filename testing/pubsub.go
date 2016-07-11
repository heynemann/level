// level game server
// https://github.com/heynemann/level
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Bernardo Heynemann <heynemann@gmail.com>

// Source file: https://github.com/nats-io/nats/blob/master/test/test.go
// Copyright 2015 Apcera Inc. All rights reserved.

package testing

import (
	gnatsdServer "github.com/nats-io/gnatsd/server"
	gnatsdTest "github.com/nats-io/gnatsd/test"
	"github.com/nats-io/nats"
)

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
