package main

import (
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"os"
)

func showList(m *Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fixCSR(w)
		writeJSON(w, m.list())
	}
}

func UIServer(addr string, m *Manager) {
	// See https://github.com/codegangsta/gin for get to known PORT environment.
	if p := os.Getenv("PORT"); p != "" {
		addr = ":" + p
	}
	http.Handle("/list", showList(m))
	http.Handle("/connect", connectByWS(m))
	http.ListenAndServe(addr, nil)
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func connectByWS(m *Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fixCSR(w)

		uuid := r.FormValue("uuid")

		if !m.Has(uuid) {
			w.WriteHeader(403)
			writeJSON(w, "Invalid magic key")
			return
		}

		ws, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			w.WriteHeader(501)
			writeJSON(w, err.Error())
			return
		}

		done := make(chan struct{})
		go wsPing(ws, done)

		c, err := NewWebSocketClientChannel(uuid, ws)
		if err != nil {
			w.WriteHeader(501)
			writeJSON(w, err.Error())
			return
		}

		go func() {
			m.Hacking(c, uuid)
			done <- struct{}{}
			ws.Close()
			log.Println("Close ws...")
		}()

		log.Println("End of request ws")
	}
}
