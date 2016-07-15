// level game server
// https://github.com/heynemann/level
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Bernardo Heynemann <heynemann@gmail.com>

package heartbeat

import (
	"time"

	"github.com/heynemann/level/extensions/pubsub"
	"github.com/heynemann/level/messaging"
)

//Service represents the heartbeat service
type Service struct {
	PubSub *pubsub.PubSub
}

//NewHeartbeatService creates a new instance of a heartbeat service
func NewHeartbeatService() *Service {
	return &Service{}
}

//HandleAction handles a given action for an user
func (p *Service) HandleAction(action *messaging.Action, reply func(*messaging.Event) error) error {
	action.Payload["serverSent"] = time.Now().UnixNano() / 1000000

	event := messaging.NewEvent("pong", action.Payload)
	reply(event)
	return nil
}

//Initialize the service
func (p *Service) Initialize(pubSub *pubsub.PubSub) {
	p.PubSub = pubSub
}
