package main

import (
	"github.com/gorilla/websocket"
	"io"
	"log"
	"net/http"
	"os/exec"
	"time"
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

func serveStatus(m *Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fixCSR(w)
		writeJSON(w, m.uuid)
	}
}

func (m *Manager) HTTPServer(f io.Reader, addr string) {
	http.HandleFunc("/tty/status", serveStatus(m))
	http.HandleFunc("/tty", serveWs(f))
	go func() {
		time.Sleep(time.Millisecond * 20)
		if true {
			log.Printf("Please open %q to see more informations\n", "http://"+addr)
		} else {
			exec.Command("xdg-open", "http://"+addr).Run()
		}
	}()
	log.Fatal(http.ListenAndServe(addr, nil))
}
