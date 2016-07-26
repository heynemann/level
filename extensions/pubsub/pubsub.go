// level game server
// https://github.com/heynemann/level
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Bernardo Heynemann <heynemann@gmail.com>

package pubsub

import (
	"fmt"
	"time"

	"github.com/heynemann/level/extensions/sessionManager"
	"github.com/heynemann/level/messaging"
	"github.com/iris-contrib/websocket"
	"github.com/nats-io/nats"
	"github.com/satori/go.uuid"
	"github.com/uber-go/zap"
)

//Player represents a player connection
type Player struct {
	SessionID string
	Socket    *websocket.Conn
	Session   *sessionManager.Session
}

//NewPlayer builds a new player instance
func NewPlayer(sessionID string, socket *websocket.Conn, session *sessionManager.Session) *Player {
	return &Player{
		SessionID: sessionID,
		Socket:    socket,
		Session:   session,
	}
}

// PubSub is responsible for handling all operations related to Publish Subscribe infrastructure
type PubSub struct {
	NatsURL          string
	Conn             *nats.EncodedConn
	SessionManager   sessionManager.SessionManager
	Logger           zap.Logger
	ConnectedPlayers map[string]*Player
	ActionTimeout    time.Duration
}

//New returns a new pubsub connection
func New(natsURL string, logger zap.Logger, manager sessionManager.SessionManager, actionTimeout time.Duration) (*PubSub, error) {
	conn, err := nats.Connect(natsURL)
	if err != nil {
		return nil, err
	}

	encoded, err := nats.NewEncodedConn(conn, nats.JSON_ENCODER)
	if err != nil {
		return nil, err
	}

	pubSub := &PubSub{
		NatsURL:          natsURL,
		Conn:             encoded,
		Logger:           logger,
		ConnectedPlayers: map[string]*Player{},
		SessionManager:   manager,
		ActionTimeout:    actionTimeout,
	}

	return pubSub, nil
}

//GetServerQueue returns the action queue for a specific server
func GetServerQueue(serverName string) string {
	return fmt.Sprintf("level.actions.server-%s", serverName)
}

//GetEventQueue returns the event queue for all servers
func GetEventQueue() string {
	return "level.events"
}

// RequestAction requests an action to a given server and returns its Event as response
func (p *PubSub) RequestAction(player *Player, action *messaging.Action, reply func(event *messaging.Event) error) error {
	var response messaging.Event
	err := p.Conn.Request(action.Key, action, &response, 5*time.Second)
	if err != nil {
		return err
	}

	err = reply(&response)
	if err != nil {
		return err
	}

	return nil
}

//RegisterPlayer registers a player to receive/send events
func (p *PubSub) RegisterPlayer(websocket *websocket.Conn) (*Player, error) {
	sessionID := uuid.NewV4().String()
	session, err := p.SessionManager.Start(sessionID)
	if err != nil {
		fmt.Println("ERROR in Session")
		return nil, err
	}
	player := NewPlayer(sessionID, websocket, session)
	p.ConnectedPlayers[sessionID] = player

	action := messaging.NewAction(sessionID, "channel.session.start", nil)
	p.RequestAction(player, action, p.getReply(websocket))

	return player, nil
}

//UnregisterPlayer removes player from connected players upon disconnection
func (p *PubSub) UnregisterPlayer(player *Player) error {
	delete(p.ConnectedPlayers, player.SessionID)
	return nil
}

func (p *PubSub) getReply(ws *websocket.Conn) func(*messaging.Event) error {
	return func(event *messaging.Event) error {
		eventJSON, err := event.MarshalJSON()
		if err != nil {
			return err
		}
		ws.WriteMessage(websocket.TextMessage, eventJSON)
		return nil
	}
}
