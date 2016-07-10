// level game server
// https://github.com/heynemann/level
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Bernardo Heynemann <heynemann@gmail.com>

package sessionManager_test

import (
	"github.com/heynemann/level/extensions/sessionManager"
	"github.com/uber-go/zap"
)

func getDefaultSM(logger zap.Logger) *sessionManager.RedisSessionManager {
	sessionManager, _ := sessionManager.GetRedisSessionManager(
		"localhost", // Redis Host
		7777,        // Redis Port
		"",          // Redis Pass
		0,           // Redis DB
		180,
		logger,
	)

	return sessionManager
}
