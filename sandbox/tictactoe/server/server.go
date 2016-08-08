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
	ServiceID  string
	RandomSeed int64
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
		action.SessionID,
		"",
		uuid.NewV4().String(),
	)

	return messaging.NewEvent(
		"tictactoe.gameplay.started",
		map[string]interface{}{
			"gameID":   game.GameID,
			"opponent": "bot",
		},
	), nil
}

func (s *GameplayService) handleMove(action *messaging.Action) (*messaging.Event, error) {
	return nil, nil
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
