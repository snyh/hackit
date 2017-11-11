package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"golang.org/x/crypto/ssh"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os/exec"
	"sync"
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
	if err != nil {
		log.Printf("Please open %q to see more informations\n", url)
	} else {
		exec.Command(bin, url).Run()
	}
}

type WriteSwitcher interface {
	io.Writer
	Switch(io.Writer)
}

type SimpleWriteSwitcher struct {
	inner io.Writer
}

func NewSimpleWriteSwitcher() WriteSwitcher {
	return &SimpleWriteSwitcher{ioutil.Discard}
}
func (p *SimpleWriteSwitcher) Write(buf []byte) (int, error) { return p.inner.Write(buf) }
func (p *SimpleWriteSwitcher) Switch(w io.Writer)            { p.inner = w }

type ChatMessage []byte

type ChatBuffer struct {
	sync.Mutex

	ws    *websocket.Conn
	index int

	ssh ssh.Channel

	buf []ChatMessage

	pending chan ChatMessage
}

func NewChatBuffer(ch ssh.Channel) *ChatBuffer {
	buf := &ChatBuffer{
		ssh: ch,
	}
	buf.WriteFromSSH([]byte("hello little fairy"))
	go buf.work()
	return buf
}

func (cb *ChatBuffer) SwitchWS(ws *websocket.Conn) {
	cb.Lock()
	cb.ws = ws
	cb.index = 0
	fmt.Println("SWTCHWS", cb.index)
	cb.Unlock()
}

func (cb *ChatBuffer) Pendings() <-chan ChatMessage {
	cb.Lock()
	defer cb.Unlock()

	n := len(cb.buf)
	np := n - cb.index

	fmt.Println("P1", n, np)
	if np <= 0 || cb.ws == nil {
		ch := make(chan ChatMessage)
		close(ch)
		return ch
	}

	ch := make(chan ChatMessage, np)

	fmt.Println("P3", np)

	for _, msg := range cb.buf[cb.index:] {
		cb.index++
		ch <- msg
	}
	close(ch)
	return ch
}

func (cb *ChatBuffer) work() {
	// TODO handling close
	for {
		time.Sleep(time.Second)
		for msg := range cb.Pendings() {
			cb.ws.WriteMessage(websocket.TextMessage, msg)
		}
	}
}

func (cb *ChatBuffer) record(msg ChatMessage) []byte {
	cb.Lock()
	cb.buf = append(cb.buf, msg)
	fmt.Println("RECORD:", string(msg))
	cb.Unlock()
	return msg
}

func (cb *ChatBuffer) WriteFromWS(msg ChatMessage) {
	fmt.Println("WriteFromWS..", string(msg))
	cb.ssh.SendRequest("chat", false, cb.record(msg))
}

func (cb *ChatBuffer) WriteFromSSH(msg ChatMessage) {
	fmt.Println("WriteFromSSH..", string(msg))
	cb.record(msg)
}
