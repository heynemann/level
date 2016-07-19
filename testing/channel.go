// level game server
// https://github.com/heynemann/level
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Bernardo Heynemann <heynemann@gmail.com>

// Source file: https://github.com/nats-io/nats/blob/master/test/test.go
// Copyright 2015 Apcera Inc. All rights reserved.

package testing

import (
	"fmt"
	"time"

	"golang.org/x/net/websocket"

	"github.com/heynemann/level/channel"
	"github.com/heynemann/level/messaging"
	"github.com/uber-go/zap"
)

////////////////////////////////////////////////////////////////////////////////
// Running gnatsd server in separate Go routines
////////////////////////////////////////////////////////////////////////////////

//RunChannel will run a server on the default port.
func RunChannel(logger zap.Logger) (*channel.Channel, error) {
	return RunChannelOnPort(3000, logger)
}

//RunChannelOnPort will run a server on the given port.
func RunChannelOnPort(port int, logger zap.Logger) (*channel.Channel, error) {
	options := channel.DefaultOptions()
	options.Port = port

	channel, err := channel.New(options, logger)
	if err != nil {
		return nil, err
	}

	go func() {
		channel.Start()
	}()
	return channel, nil
}

//TestConnection to a channel
type TestConnection struct {
	Channel *channel.Channel
	ws      *websocket.Conn
}

//NewChannelTestConnection creates a new test connection to a channel
func NewChannelTestConnection(channel *channel.Channel) (*TestConnection, error) {
	tc := &TestConnection{
		Channel: channel,
	}
	origin := fmt.Sprintf("http://%s", channel.ServerOptions.Host)
	url := fmt.Sprintf("ws://%s:%d", channel.ServerOptions.Host, channel.ServerOptions.Port)

	var err error
	var ws *websocket.Conn

	for {
		ws, err = websocket.Dial(url, "", origin)
		if err != nil {
			time.Sleep(10 * time.Millisecond)
			continue
		}
		break
	}
	tc.ws = ws

	return tc, nil
}

//Send an action through the channel
func (tc *TestConnection) Send(action *messaging.Action) error {
	payload, _ := action.MarshalJSON()
	_, err := tc.ws.Write(payload)
	if err != nil {
		return err
	}

	return nil
}

//Receive the next event
func (tc *TestConnection) Receive(to ...time.Duration) (*messaging.Event, error) {
	timeout := 100 * time.Millisecond
	if to != nil && len(to) == 1 {
		timeout = to[0]
	}
	var msg = make([]byte, 10240)
	var err error
	var n int
	tc.ws.SetReadDeadline(time.Now().Add(timeout))
	if n, err = tc.ws.Read(msg); err != nil {
		return nil, err
	}
	if n == 0 {
		return nil, fmt.Errorf("Could not read any bytes from the socket.")
	}
	event := &messaging.Event{}
	err = event.UnmarshalJSON(msg[:n])
	if err != nil {
		return nil, err
	}

	return event, nil
}
