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
)

func main() {
	s, _ := sessionManager.GetSessionManager("localhost", 7777, "", 0, 180)
	fmt.Println("Hello!", s)
}
