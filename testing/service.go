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
func RunService(channelPort int, service registry.Service, logger zap.Logger, configPath string) (*channel.Channel, *service.Server, error) {
	options := channel.DefaultOptions()
	options.ConfigFile = configPath

	channel, err := RunChannelWithOptions(options, logger)
	if err != nil {
		return nil, nil, err
	}
	server, err := StartService(service, logger, configPath)
	if err != nil {
		return nil, nil, err
	}

	return channel, server, nil
}

//StartService and listen in goroutine
func StartService(serv registry.Service, logger zap.Logger, configPath string) (*service.Server, error) {
	details := serv.GetServiceDetails()
	l := logger.With(
		zap.String("serverName", details.Name),
		zap.String("serverDescription", details.Description),
		zap.String("serverVersion", details.Version),
		zap.String("serverID", details.ServiceID),
	)

	server, err := service.NewServer(serv, l, configPath)
	if err != nil {
		return nil, err
	}

	go func() {
		server.Logger.Debug("Running backend server")
		server.Listen()
	}()

	return server, nil
}
