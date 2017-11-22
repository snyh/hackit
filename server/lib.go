package main

import (
	"encoding/json"
	"fmt"
	"github.com/glycerine/rbuf"
	"github.com/gorilla/websocket"
	"golang.org/x/crypto/ssh"
	"io"
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

type ChannelHistory struct {
	sync.Mutex

	ws *websocket.Conn

	ssh  ssh.Channel
	ring *rbuf.FixedSizeRingBuf

	done bool
}

func (cb *ChannelHistory) Switch(ws *websocket.Conn) {
	cb.Lock()
	cb.ws = ws
	cb.Unlock()
}

func NewChannelHistory(ssh ssh.Channel) *ChannelHistory {
	cb := &ChannelHistory{
		ssh:  ssh,
		ring: rbuf.NewFixedSizeRingBuf(1024 * 1024 * 20), // Max tty history size 20MB
	}
	go cb.Work()
	return cb
}
func (cb *ChannelHistory) Write(p []byte) (int, error) {
	n, err := cb.ssh.Write(p)
	if err != nil {
		return n, err
	}
	return cb.ring.Write(p)
}
func (cb *ChannelHistory) Read(p []byte) (int, error) {
	return cb.ssh.Read(p)
}
func (cb *ChannelHistory) Close() error {
	cb.Lock()
	cb.done = true
	if cb.ws != nil {
		cb.ws.Close()
	}
	cb.Unlock()
	return cb.ssh.Close()
}

func (cb *ChannelHistory) Work() {
	for {
		if cb.done {
			return
		}
		time.Sleep(time.Millisecond * 200)
		cb.Lock()

		if cb.ws == nil {
			cb.Unlock()
			continue
		}
		cb.Unlock()

		_, err := io.Copy(wsWrap{cb.ws}, cb.ring)
		if err != nil && err != io.EOF {
			fmt.Println("CB WORK FAIL:", err)
		}
	}
}

func openUrl(url string) {
	bin, err := exec.LookPath("xdg-open")
	if err != nil {
		log.Printf("Please open %q to see more informations\n", url)
	} else {
		exec.Command(bin, url).Run()
	}
}

type ChatMessage struct {
	Author string      `json:"author"`
	Type   string      `json:"type"`
	Data   interface{} `json:"data"`
}

func (msg ChatMessage) Marshal() []byte {
	bs, _ := json.Marshal(msg)
	return bs
}

func (msg ChatMessage) Invert() ChatMessage {
	if msg.Author == "me" {
		msg.Author = "theme"
	} else {
		msg.Author = "me"
	}
	return msg
}

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
	buf.work()
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

	if np <= 0 || cb.ws == nil {
		ch := make(chan ChatMessage)
		close(ch)
		return ch
	}

	ch := make(chan ChatMessage, np)

	for _, msg := range cb.buf[cb.index:] {
		cb.index++
		ch <- msg
	}
	close(ch)
	return ch
}

func (cb *ChatBuffer) work() {
	// TODO handling close
	done := false
	go func() {
		for {
			time.Sleep(time.Millisecond * 150)
			for msg := range cb.Pendings() {
				cb.ws.WriteMessage(websocket.TextMessage, msg.Marshal())
			}
			if done {
				return
			}
		}
	}()
	go func() {
		for {
			if cb.ws != nil {
				_, bs, err := cb.ws.ReadMessage()
				if err != nil {
					done = true
					return
				}
				cb.WriteFromWS(bs)
			}
			time.Sleep(time.Millisecond * 100)
		}
	}()
}

func (cb *ChatBuffer) record(bs []byte) ChatMessage {
	cb.Lock()
	var msg ChatMessage
	json.Unmarshal(bs, &msg)
	cb.buf = append(cb.buf, msg)
	cb.Unlock()
	return msg
}

func (cb *ChatBuffer) WriteFromWS(bs []byte) {
	fmt.Println("WriteFromWS..", string(bs))
	msg := cb.record(bs)
	cb.ssh.SendRequest("chat", false, msg.Invert().Marshal())
}

func (cb *ChatBuffer) WriteFromSSH(msg []byte) {
	fmt.Println("WriteFromSSH..", string(msg))
	cb.record(msg)
}
