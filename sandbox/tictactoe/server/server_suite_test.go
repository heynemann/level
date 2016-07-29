// level game server
// https://github.com/heynemann/level
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Bernardo Heynemann <heynemann@gmail.com>

package tictactoe_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestTicTacToe(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Tic-Tac-Toe Suite")
}
