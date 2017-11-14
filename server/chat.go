package main

import (
	"github.com/gorilla/websocket"
	"net/http"
)

// ServeChat 收发WebSocket上的chat message 到c.chatQueue上
func (c *HackItConn) ServeChat(w http.ResponseWriter, r *http.Request) {
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}

	// setup websocket
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		writeJSON(w, 501, err)
		return
	}

	c.chatBuffer.SwitchWS(ws)
}
