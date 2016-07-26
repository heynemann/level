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
	"net/url"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"github.com/heynemann/level/channel"
	"github.com/heynemann/level/messaging"
	"github.com/uber-go/zap"
)

// DefaultTestOptions returns default options for channel tests
func DefaultTestOptions() *channel.Options {
	return channel.NewOptions(
		"0.0.0.0",
		3000,
		true,
		"../config/test.yaml",
	)
}

//RunChannel will run a server on the default port.
func RunChannel(logger zap.Logger) (*channel.Channel, error) {
	return RunChannelOnPort(3000, logger)
}

//RunChannelOnPort will run a server on the given port.
func RunChannelOnPort(port int, logger zap.Logger) (*channel.Channel, error) {
	options := DefaultTestOptions()
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
	Channel    *channel.Channel
	ws         *websocket.Conn
	Received   []*messaging.Event
	Errors     []error
	Stop       chan bool
	PingTicker *time.Ticker
	Waited     int
}

//NewChannelTestConnection creates a new test connection to a channel
func NewChannelTestConnection(channel *channel.Channel) (*TestConnection, error) {
	timeout := time.Duration(100 * time.Millisecond)
	pongWait := 60 * time.Millisecond

	tc := &TestConnection{
		Channel:    channel,
		Received:   []*messaging.Event{},
		Errors:     []error{},
		Stop:       make(chan bool),
		PingTicker: time.NewTicker(50 * time.Millisecond),
	}
	wsURL := fmt.Sprintf("%s:%d", channel.ServerOptions.Host, channel.ServerOptions.Port)

	u := url.URL{Scheme: "ws", Host: wsURL, Path: "/"}

	start := time.Now()

	var ws *websocket.Conn
	var err error

	for {
		if time.Now().Sub(start) > timeout {
			return nil, fmt.Errorf("Timed out trying to establish websocket connection to %s.", wsURL)
		}

		ws, _, err = websocket.DefaultDialer.Dial(u.String(), nil)
		if err != nil {
			time.Sleep(5 * time.Millisecond)
			continue
		}
		break
	}
	tc.ws = ws

	ws.SetPongHandler(func(string) error { ws.SetReadDeadline(time.Now().Add(pongWait)); return nil })

	go func(conn *TestConnection) {
		for {
			select {
			case <-conn.PingTicker.C:
				if err := ws.WriteMessage(websocket.PingMessage, []byte{}); err != nil {
					return
				}
			case <-conn.Stop:
				return
			default:
				ev, err := conn.Receive()
				if err != nil {
					conn.Errors = append(conn.Errors, err)
					if strings.HasSuffix(err.Error(), "i/o timeout") {
						close(conn.Stop)
						return
					}
				}
				conn.Received = append(conn.Received, ev)
			}
		}
	}(tc)

	return tc, nil
}

// Close websocket connection and stop listening for events.
func (tc *TestConnection) Close() {
	close(tc.Stop)
	tc.ws.Close()
}

//Send an action through the channel
func (tc *TestConnection) Send(action *messaging.Action) error {
	payload, _ := action.MarshalJSON()
	err := tc.ws.WriteMessage(websocket.TextMessage, payload)
	if err != nil {
		return err
	}

	return nil
}

//WaitFor messages to come
func (tc *TestConnection) WaitFor(messages int, to ...time.Duration) error {
	timeout := 50 * time.Millisecond
	if to != nil && len(to) == 1 {
		timeout = to[0]
	}

	start := time.Now()

	for {
		if time.Now().Sub(start) > timeout {
			return fmt.Errorf("Timed out waiting for WebSocket messages to come in.")
		}

		if len(tc.Received) >= messages {
			tc.Waited += messages
			return nil
		}

		time.Sleep(5 * time.Millisecond)
	}
}

//Receive the next event
func (tc *TestConnection) Receive(to ...time.Duration) (*messaging.Event, error) {
	c := tc.ws

	timeout := 10 * time.Millisecond
	if to != nil && len(to) == 1 {
		timeout = to[0]
	}
	c.SetReadDeadline(time.Now().Add(timeout))

	_, message, err := c.ReadMessage()
	if err != nil {
		return nil, err
	}
	ev := messaging.Event{}
	err = ev.UnmarshalJSON(message)
	if err != nil {
		return nil, err
	}

	return &ev, nil
}
