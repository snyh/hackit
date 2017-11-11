package main

import (
	"fmt"
	"github.com/gorilla/websocket"
	"golang.org/x/crypto/ssh"
	"net/http"
	"sync"
	"time"
)

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

	done := make(chan struct{})

	c.chatBuffer.SwitchWS(ws)

	go func() {
		for {
			_, msg, err := ws.ReadMessage()
			if err != nil {
				done <- struct{}{}
				return
			}
			c.chatBuffer.WriteFromWS(msg)

			time.Sleep(time.Millisecond * 100)
		}
	}()

	go func() {
		wsPing(ws, done)
		ws.Close()
	}()
}
