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
	"github.com/kataras/iris/websocket"
	"github.com/nats-io/nats"
	"github.com/satori/go.uuid"
	"github.com/uber-go/zap"
)

//Player represents a player connection
type Player struct {
	SessionID string
	Socket    websocket.Connection
	Session   *sessionManager.Session
}

//NewPlayer builds a new player instance
func NewPlayer(sessionID string, socket websocket.Connection, session *sessionManager.Session) *Player {
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

// SubscribeActions subscribes a specific server to all actions arriving in its queue
//func (p *PubSub) SubscribeActions(serverName string, callback func(func(*messaging.Event), *messaging.Action)) error {
//p.Conn.Subscribe(GetServerQueue(serverName), func(subj, reply string, action *messaging.Action) {
//replyFunc := func(e *messaging.Event) {
//p.Conn.Publish(reply, e)
//}
//callback(replyFunc, action)
//})
//return nil
//}

// RequestAction requests an action to a given server and returns its Event as response
func (p *PubSub) RequestAction(player *Player, action *messaging.Action, reply func(event *messaging.Event) error, serverReceived int64) error {
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
func (p *PubSub) RegisterPlayer(websocket websocket.Connection) error {
	sessionID := uuid.NewV4().String()
	session, err := p.SessionManager.Start(sessionID)
	if err != nil {
		fmt.Println("ERROR in Session")
		return err
	}
	player := NewPlayer(sessionID, websocket, session)
	p.ConnectedPlayers[sessionID] = player

	p.BindEvents(websocket, player)
	return nil
}

//UnregisterPlayer removes player from connected players upon disconnection
func (p *PubSub) UnregisterPlayer(player *Player) error {
	delete(p.ConnectedPlayers, player.SessionID)
	return nil
}

func (p *PubSub) getReply(websocket websocket.Connection) func(*messaging.Event) error {
	return func(event *messaging.Event) error {
		eventJSON, err := event.MarshalJSON()
		if err != nil {
			websocket.EmitError(fmt.Sprintf("Failed to process action: %s", err.Error()))
			return err
		}
		websocket.EmitMessage(eventJSON)
		return nil
	}
}

//BindEvents listens to websocket events.
func (p *PubSub) BindEvents(websocket websocket.Connection, player *Player) {
	websocket.OnMessage(func(message []byte) {
		received := time.Now().UnixNano()

		var action messaging.Action
		err := action.UnmarshalJSON(message)
		if err != nil {
			return
		}

		err = p.RequestAction(player, &action, p.getReply(websocket), received)
		if err != nil {
			fmt.Println("GOT AN ERROR REQUESTING ACTION: ", err.Error())
		}
	})
	// to all except this connection ->
	//c.To(websocket.Broadcast).Emit("chat", "Message from: "+c.ID()+"-> "+message)

	// to the client ->
	//c.Emit("chat", "Message from myself: "+message)

	//send the message to the whole room,
	//all connections are inside this room will receive this message
	//c.Emit("chat", "From: "+c.ID()+": "+message)
	//})

	websocket.OnDisconnect(func() {
		p.UnregisterPlayer(player)
	})
}
