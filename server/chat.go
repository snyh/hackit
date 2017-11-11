package main

import (
	"fmt"
	"github.com/gorilla/websocket"
	"golang.org/x/crypto/ssh"
	"net/http"
	"time"
)

func makeChatRobot(server ssh.Channel) error {
	go func() {
		for {
			<-time.After(time.Second)
			server.SendRequest("chat", false, []byte(fmt.Sprintf("Server Time is : %s", time.Now())))
		}
	}()
	return nil
}

type ChatMessage []byte

type ChatQueue struct {
	In  chan<- ChatMessage
	Out <-chan ChatMessage
}

func (c *HackItConn) onReceiveChat(bs ChatMessage) {
	c.chatQueue.In <- bs
}

func (c *HackItConn) chatForawrd() {
	in := make(chan ChatMessage, 1)
	out := make(chan ChatMessage, 1)
	c.chatQueue = ChatQueue{in, out}

	// handle in message
	go func() {
		for msg := range in {
			fmt.Printf("\033[31m %s \033[0m\n", msg)
		}
	}()

	// handle out message
	go func() {
	}()
}

// ServeChat 将本地websockt的chat message一一映射到c.ssh中去
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
	go func() {
		// io.Copy(wsWrap{ws}, rp)
		done <- struct{}{}
		ws.Close()
	}()
	go wsPing(ws, done)
}

type wsChatWrap struct {
	core *websocket.Conn
}

func (w wsChatWrap) Write(p []byte) (int, error) {
	return len(p), w.core.WriteMessage(websocket.TextMessage, p)
}
func (w wsChatWrap) Read(p []byte) (int, error) {
	_, bs, err := w.core.ReadMessage()
	copy(p, bs)
	return len(bs), err
}
func (w wsChatWrap) Close() error {
	return w.core.Close()
}
