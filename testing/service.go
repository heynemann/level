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
	"github.com/heynemann/level/channel"
	"github.com/heynemann/level/extensions/serviceRegistry"
	"github.com/heynemann/level/services"
	"github.com/uber-go/zap"
)

//RunService will run a service and a channel on the default port.
func RunService(channelPort int, service registry.Service, logger zap.Logger) (*channel.Channel, *service.Server, error) {
	channel, err := RunChannelOnPort(channelPort, logger)
	if err != nil {
		return nil, nil, err
	}
	server, err := StartService(service)
	if err != nil {
		return nil, nil, err
	}

	return channel, server, nil
}

//StartService and listen in goroutine
func StartService(serv registry.Service) (*service.Server, error) {
	server, err := service.NewServer(serv)
	if err != nil {
		return nil, err
	}

	go func() {
		server.Listen()
	}()

	return server, nil
}
