// level game server
// https://github.com/heynemann/level
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Bernardo Heynemann <heynemann@gmail.com>

package channel

import (
	"fmt"
	"strings"

	"github.com/heynemann/level/messaging"
	"github.com/iris-contrib/websocket"
	"github.com/kataras/iris"
)

//WebsocketHandlerInstance is an instance of a websocket connection
type WebsocketHandlerInstance struct {
	channel *Channel
}

func (ws *WebsocketHandlerInstance) handleWebSocket(conn *websocket.Conn) {
	defer conn.Close()

	c := ws.channel

	player, err := c.PubSub.RegisterPlayer(conn)
	if err != nil {
		fmt.Println("Error registering player: ", err)
		return
	}

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			if strings.Contains(err.Error(), "abnormal closure") {
				return
			}
			fmt.Println("read:", err)
			continue
		}

		action := &messaging.Action{}
		err = action.UnmarshalJSON(message)
		if err != nil {
			fmt.Println("Error with action: ", err)
			continue
		}

		c.PubSub.RequestAction(player, action, func(event *messaging.Event) error {
			json, err := event.MarshalJSON()
			if err != nil {
				fmt.Println("Error sending action reply")
				return err
			}
			conn.WriteMessage(websocket.TextMessage, json)
			return nil
		})
	}
}

//WebSocketHandler handles websocket connections
func WebSocketHandler(c *Channel) func(*iris.Context) {
	ws := &WebsocketHandlerInstance{
		channel: c,
	}
	upgrader := websocket.New(ws.handleWebSocket)
	upgrader.DontCheckOrigin()

	return func(ctx *iris.Context) {
		err := upgrader.Upgrade(ctx)
		if err != nil {
			fmt.Println("Error in connection: ", err)
		}
	}
}
