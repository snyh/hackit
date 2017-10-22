package main

import (
	"encoding/json"
	"github.com/gorilla/websocket"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os/exec"
	"time"
)

func wsPing(ws *websocket.Conn, done <-chan struct{}) {
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
				log.Println("ws ping:", err)
			}
		case <-done:
			return
		}
	}
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
func (w wsWrap) Close() error {
	return w.core.Close()
}

func fixCSR(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
}
func writeJSON(w http.ResponseWriter, code int, v interface{}) {
	bs, err := json.Marshal(v)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(code)
	w.Write(bs)
}

func openUrl(url string) {
	bin, err := exec.LookPath("xdg-open")
	if err != nil || true {
		log.Printf("Please open %q to see more informations\n", url)
	} else {
		exec.Command(bin, url).Run()
	}
}

type WriteSwitcher interface {
	io.Writer
	Switch(io.Writer)
}
type SimpleSwitcher struct {
	inner io.Writer
}

func NewSimpleSwitcher() WriteSwitcher {
	return &SimpleSwitcher{ioutil.Discard}
}
func (p *SimpleSwitcher) Write(buf []byte) (int, error) { return p.inner.Write(buf) }
func (p *SimpleSwitcher) Switch(w io.Writer)            { p.inner = w }
