// level game server
// https://github.com/heynemann/level
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Bernardo Heynemann <heynemann@gmail.com>

package main

import (
	"fmt"

	"github.com/heynemann/level/extensions/sessionManager"
	"github.com/uber-go/zap"
)

func main() {
	logger := zap.NewJSON()
	s, _ := sessionManager.GetRedisSessionManager("localhost", 7777, "", 0, 180, logger)
	fmt.Println("Hello!", s)
}
