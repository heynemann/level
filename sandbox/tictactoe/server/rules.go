// level game server
// https://github.com/heynemann/level
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Bernardo Heynemann <heynemann@gmail.com>

package tictactoe

import (
	"fmt"
	"math/rand"
)

//Game of tic-tac-toe
type Game struct {
	AgainstBot       bool
	GameID           string
	Player1SessionID string
	Player2SessionID string
	Board            *Board
}

//NewGame with the specified players
func NewGame(againstBot bool, gameID, player1SessionID, player2SessionID string) *Game {
	board := &Board{
		Pieces: [][]int{
			[]int{0, 0, 0},
			[]int{0, 0, 0},
			[]int{0, 0, 0},
		},
		CurrentPlayer: 1,
	}

	return &Game{
		AgainstBot:       againstBot,
		Board:            board,
		GameID:           gameID,
		Player1SessionID: player1SessionID,
		Player2SessionID: player2SessionID,
	}
}

//Board represents a 3x3 game of tic-tac-toe
type Board struct {
	CurrentPlayer int
	Pieces        [][]int
}

//Winner indicates whether there's a winner for this board?
func (b *Board) Winner() int {
	pieces := b.Pieces

	//cols
	for x := 0; x < 3; x++ {
		if pieces[x][0] != 0 && pieces[x][0] == pieces[x][1] && pieces[x][0] == pieces[x][2] {
			return pieces[x][0]
		}
	}

	//rows
	for y := 0; y < 3; y++ {
		if pieces[0][y] != 0 && pieces[0][y] == pieces[1][y] && pieces[0][y] == pieces[2][y] {
			return pieces[0][y]
		}
	}

	//diags
	if pieces[0][0] != 0 && pieces[0][0] == pieces[1][1] && pieces[0][0] == pieces[2][2] {
		return pieces[0][0]
	}
	if pieces[2][0] != 0 && pieces[2][0] == pieces[1][1] && pieces[2][0] == pieces[0][2] {
		return pieces[2][0]
	}

	return 0
}

//IsGameOver for this board?
func (b *Board) IsGameOver() bool {
	return b.Winner() != 0 || b.IsDraw()
}

//IsDraw for this board?
func (b *Board) IsDraw() bool {
	return b.Winner() == 0 && b.GetAvailableMoves() == 0
}

//GetAvailableMoves in the board
func (b *Board) GetAvailableMoves() int {
	pieces := b.Pieces
	moves := 0

	for x := 0; x < 3; x++ {
		for y := 0; y < 3; y++ {
			if pieces[x][y] == 0 {
				moves++
			}
		}
	}
	return moves
}

//GetBotMove for next valid position
func (b *Board) GetBotMove() (int, int) {
	pieces := b.Pieces

	for {
		rx := rand.Intn(3)
		ry := rand.Intn(3)

		if pieces[rx][ry] == 0 {
			return int(rx), int(ry)
		}
	}
}

func (b *Board) validateMove(player, posX, posY int) (string, bool) {
	if b.CurrentPlayer != player {
		return "Not player's move!!!", false
	}
	if b.Pieces[posX][posY] != 0 {
		return fmt.Sprintf("Position %d:%d is not empty!\n", posX, posY), false
	}

	if b.IsGameOver() {
		return "Game over!!!", false
	}

	return "", true
}

//TickAs a specific player in a specific position.
func (b *Board) TickAs(player, posX, posY int) error {
	msg, isValid := b.validateMove(player, posX, posY)
	if !isValid {
		return fmt.Errorf(msg)
	}

	if player == 1 {
		b.CurrentPlayer = 2
	} else {
		b.CurrentPlayer = 1
	}
	b.Pieces[posX][posY] = player

	return nil
}
