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

	"github.com/heynemann/level/messaging"
	"github.com/nats-io/nats"
	"github.com/uber-go/zap"
)

// PubSub is responsible for handling all operations related to Publish Subscribe infrastructure
type PubSub struct {
	NatsURL string
	Conn    *nats.EncodedConn
	Logger  zap.Logger
}

//New returns a new pubsub connection
func New(natsURL string) (*PubSub, error) {
	conn, err := nats.Connect(natsURL)
	if err != nil {
		return nil, err
	}

	encoded, err := nats.NewEncodedConn(conn, nats.JSON_ENCODER)
	if err != nil {
		return nil, err
	}

	pubSub := PubSub{
		NatsURL: natsURL,
		Conn:    encoded,
	}

	return &pubSub, nil
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
func (p *PubSub) SubscribeActions(serverName string, callback func(func(*messaging.Event), *messaging.Action)) error {
	p.Conn.Subscribe(GetServerQueue(serverName), func(subj, reply string, action *messaging.Action) {
		replyFunc := func(e *messaging.Event) {
			p.Conn.Publish(reply, e)
		}
		callback(replyFunc, action)
	})
	return nil
}

// RequestAction requests an action to a given server and returns its Event as response
func (p *PubSub) RequestAction(serverName string, action *messaging.Action, timeout time.Duration) (*messaging.Event, error) {
	var response messaging.Event
	err := p.Conn.Request(GetServerQueue(serverName), action, &response, timeout)
	if err != nil {
		return nil, err
	}

	return &response, nil
}

// SubscribeEvents subscribes to all events arriving from the servers
func (p *PubSub) SubscribeEvents(callback func(*messaging.Event)) error {
	p.Conn.Subscribe(GetEventQueue(), callback)
	return nil
}

// PublishEvent publishes an event to all the channels
func (p *PubSub) PublishEvent(event *messaging.Event) error {
	p.Conn.Publish(GetEventQueue(), event)
	return nil
}
