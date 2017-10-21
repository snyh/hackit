package main

import (
	"github.com/gorilla/websocket"
	"io"
	"log"
	"net/http"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func serveWs(f io.Reader) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ws, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println("upgrade:", err)
			return
		}

		done := make(chan struct{})
		go func() {
			io.Copy(wsWrap{ws}, f)
			done <- struct{}{}
			ws.Close()
		}()
		go wsPing(ws, done)
	}
}
