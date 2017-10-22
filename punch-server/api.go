package main

import (
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
)

func showList(m *Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fixCSR(w)
		writeJSON(w, 200, m.list())
	}
}

type ReactRouter struct {
	fs    http.Handler
	other http.Handler
}

func (rr ReactRouter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	p := r.URL.Path

	if strings.HasPrefix(r.URL.Path, "/ws") {
		rr.other.ServeHTTP(w, r)
		return
	}

	if _, err := os.Stat("./ui/build/" + p); err != nil {
		bs, _ := ioutil.ReadFile("./ui/build/index.html")
		w.WriteHeader(200)
		w.Write(bs)
		return
	} else {
		rr.fs.ServeHTTP(w, r)
	}
}

func UIServer(addr string, m *Manager) {
	// See https://github.com/codegangsta/gin for get to known PORT environment.
	if p := os.Getenv("PORT"); p != "" {
		addr = ":" + p
	}

	r := mux.NewRouter()
	r.HandleFunc("/ws", connectByWS(m))

	http.ListenAndServe(addr, ReactRouter{
		fs:    http.FileServer(http.Dir("./ui/build/")),
		other: r,
	})
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func connectByWS(m *Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fixCSR(w)

		uuid := r.FormValue("uuid")

		if !m.Has(uuid) {
			writeJSON(w, 403, "Invalid magic key")
			return
		}

		ws, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			writeJSON(w, 501, err.Error())
			return
		}

		done := make(chan struct{})
		go wsPing(ws, done)

		c, err := NewWebSocketClientChannel(uuid, ws)
		if err != nil {
			writeJSON(w, 501, err.Error())
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
