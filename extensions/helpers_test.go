// level game server
// https://github.com/heynemann/level
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Bernardo Heynemann <heynemann@gmail.com>

package extensions_test

import "github.com/heynemann/level/extensions"

func getDefaultSM() *extensions.SessionManager {
	sessionManager, _ := extensions.GetSessionManager(
		"localhost", // Redis Host
		7777,        // Redis Port
		"",          // Redis Pass
		0,           // Redis DB
	)

	return sessionManager
}
