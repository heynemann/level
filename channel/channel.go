// level game server
// https://github.com/heynemann/level
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Bernardo Heynemann <heynemann@gmail.com>

package channel

import "github.com/uber-go/zap"

//Heartbeat extension responsible for service registry for all backend servers
type Channel struct {
	Logger zap.Logger
}

//New opens a new channel connection
func New(logger zap.Logger) (*Channel, error) {
	l := logger.With(
		zap.String("source", "channel"),
	)
	c := Channel{
		Logger: l,
	}

	return &c, nil
}
