// level game server
// https://github.com/heynemann/level
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Bernardo Heynemann <heynemann@gmail.com>

package main

import (
	"github.com/heynemann/level/sandbox/tictactoe/server"
	"github.com/heynemann/level/services"
	"github.com/satori/go.uuid"
	"github.com/uber-go/zap"
)

func main() {
	serv := &tictactoe.GameplayService{
		ServiceID: uuid.NewV4().String(),
	}
	logger := zap.NewJSON(zap.InfoLevel)
	service.Run(serv, logger, "../../config/default.yaml")
}
