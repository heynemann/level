// level game server
// https://github.com/heynemann/level
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Bernardo Heynemann <heynemann@gmail.com>

package tictactoe

import (
	"fmt"

	"github.com/heynemann/level/extensions/serviceRegistry"
	"github.com/heynemann/level/messaging"
	"github.com/heynemann/level/metadata"
	"github.com/satori/go.uuid"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

//GameplayService for tic-tac-toe
type GameplayService struct {
	ServiceID string
	Games     map[string]*Game
}

//NewGameplayService returns a configured gameplay service.
func NewGameplayService(serviceID string) *GameplayService {
	service := &GameplayService{
		ServiceID: serviceID,
		Games:     map[string]*Game{},
	}

	return service
}

//GetServiceID for this service
func (s *GameplayService) GetServiceID() string {
	return s.ServiceID
}

//HandleAction for the game
func (s *GameplayService) HandleAction(subject string, action *messaging.Action) (*messaging.Event, error) {
	switch action.Key {
	case "tictactoe.gameplay.start":
		return s.handleMatchmaking(action)
	case "tictactoe.gameplay.move":
		return s.handleMove(action)
	default:
		return nil, fmt.Errorf("Cannot process action identified by: %s", action.Key)
	}
}

func (s *GameplayService) handleMatchmaking(action *messaging.Action) (*messaging.Event, error) {
	game := NewGame(
		true,
		uuid.NewV4().String(),
		action.SessionID,
		"bot",
	)

	s.Games[game.GameID] = game

	return messaging.NewEvent(
		"tictactoe.gameplay.started",
		map[string]interface{}{
			"gameID":   game.GameID,
			"opponent": "bot",
		},
	), nil
}

func asInt(value interface{}) int {
	return int(value.(float64))
}

func (s *GameplayService) handleMove(action *messaging.Action) (*messaging.Event, error) {
	actionData := action.Payload.(map[string]interface{})
	gameID := actionData["gameID"].(string)
	game, ok := s.Games[gameID]
	if !ok {
		err := fmt.Errorf("Failed to retrieve game with id %s", gameID)
		return messaging.NewEvent(
			"tictactoe.gameplay.invalid-move",
			map[string]interface{}{
				"gameID": game.GameID,
				"error":  err.Error(),
			},
		), nil
	}

	b := game.Board

	player := 1
	if action.SessionID == game.Player2SessionID {
		player = 2
	}

	err := b.TickAs(player, asInt(actionData["posX"]), asInt(actionData["posY"]))
	if err != nil {
		return messaging.NewEvent(
			"tictactoe.gameplay.invalid-move",
			map[string]interface{}{
				"gameID": game.GameID,
				"error":  err.Error(),
			},
		), nil
	}

	if b.IsGameOver() {
		//Remove the game from the current games
		delete(s.Games, gameID)

		if b.IsDraw() {
			return messaging.NewEvent(
				"tictactoe.gameplay.result",
				map[string]interface{}{
					"gameID": game.GameID,
					"board":  b.Pieces,
					"winner": nil,
				},
			), nil
		}

		return messaging.NewEvent(
			"tictactoe.gameplay.result",
			map[string]interface{}{
				"gameID": game.GameID,
				"board":  b.Pieces,
				"winner": b.Winner(),
			},
		), nil
	}

	if game.AgainstBot {
		posX, posY := b.GetBotMove()
		b.TickAs(2, posX, posY)
	}

	return messaging.NewEvent(
		"tictactoe.gameplay.status",
		map[string]interface{}{
			"gameID": game.GameID,
			"board":  b.Pieces,
		},
	), nil
}

//SetServerFlags adds flags to when this service is run
func (s *GameplayService) SetServerFlags(cmd *cobra.Command) {}

//SetDefaultConfigurations sets configuration defaults if they are not found in config file
func (s *GameplayService) SetDefaultConfigurations(config *viper.Viper) {}

//GetServiceDetails ditto
func (s *GameplayService) GetServiceDetails() *registry.ServiceDetails {
	return registry.NewServiceDetails(
		s.ServiceID,
		"tictactoe.gameplay",
		"tictactoe",
		"Play tic-tac-toe with friends.",
		metadata.VERSION,
		false,
	)
}
