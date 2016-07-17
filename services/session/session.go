// level game server
// https://github.com/heynemann/level
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Bernardo Heynemann <heynemann@gmail.com>

package session

import (
	"fmt"
	"strings"

	"github.com/heynemann/level/extensions/pubsub"
	"github.com/heynemann/level/extensions/sessionManager"
	"github.com/heynemann/level/messaging"
)

//Service represents the heartbeat service
type Service struct {
	PubSub         *pubsub.PubSub
	SessionManager sessionManager.SessionManager
}

//NewSessionService creates a new instance of a heartbeat service
func NewSessionService() *Service {
	return &Service{}
}

//ShouldHandleAction identifies whether this service should handle the incoming action
func (p *Service) ShouldHandleAction(action *messaging.Action) bool {
	return strings.HasPrefix(action.Key, "channel.session")
}

//HandleSessionRejoin happens when a player wants to rejoin the channel
func (p *Service) HandleSessionRejoin(sessionID string, action *messaging.Action, reply func(*messaging.Event) error) error {
	return nil
}

//HandleSessionStart happens when a player first joins the channel
func (p *Service) HandleSessionStart(sessionID string, action *messaging.Action, reply func(*messaging.Event) error) error {
	event := messaging.NewEvent("channel.session.joined", map[string]interface{}{"sessionID": sessionID})
	err := reply(event)
	if err != nil {
		return err
	}

	return nil
}

//HandleAction handles a given action for a player
func (p *Service) HandleAction(sessionID string, action *messaging.Action, reply func(*messaging.Event) error, serverReceived int64) error {
	switch action.Key {
	case "channel.session.start":
		return p.HandleSessionStart(sessionID, action, reply)
	case "channel.session.rejoin":
		return p.HandleSessionRejoin(sessionID, action, reply)
	default:
		return fmt.Errorf("Cannot process action idenfied by: %s", action.Key)
	}
}

//Initialize the service
func (p *Service) Initialize(pubSub *pubsub.PubSub) {
	p.PubSub = pubSub
	p.SessionManager = p.PubSub.SessionManager
}
