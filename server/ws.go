package main

import (
	"encoding/json"
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

type wsWrap struct {
	core *websocket.Conn
}

func (w wsWrap) Write(p []byte) (int, error) {
	return len(p), w.core.WriteMessage(websocket.TextMessage, p)
}
func (w wsWrap) Read(p []byte) (int, error) {
	_, bs, err := w.core.ReadMessage()
	copy(p, bs)
	return len(bs), err
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

		go func() {
			const (
				// Time allowed to write a message to the peer.
				writeWait = 10 * time.Second

				// Time allowed to read the next pong message from the peer.
				pongWait = 60 * time.Second

				// Send pings to peer with this period. Must be less than pongWait.
				pingPeriod = (pongWait * 9) / 10
			)

			ticker := time.NewTicker(pingPeriod)
			defer ticker.Stop()
			for {
				select {
				case <-ticker.C:
					if err := ws.WriteControl(websocket.PingMessage, []byte{}, time.Now().Add(writeWait)); err != nil {
						log.Println("ping:", err)
					}
				case <-done:
					return
				}
			}
		}()
	}
}

func fixCSR(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
}
func writeJSON(w http.ResponseWriter, v interface{}) {
	bs, err := json.Marshal(v)
	if err != nil {
		w.WriteHeader(501)
		return
	}
	w.Write(bs)
}

func serveStatus(m *Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fixCSR(w)
		writeJSON(w, m.uuid)
	}
}

func (m *Manager) UIServer(f io.Reader, addr string) {
	http.HandleFunc("/tty/status", serveStatus(m))
	http.HandleFunc("/tty", serveWs(f))
	http.Handle("/", http.FileServer(http.Dir("./ui/build")))
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
