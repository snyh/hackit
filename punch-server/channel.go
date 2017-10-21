package main

import (
	"fmt"
	"github.com/gorilla/websocket"
	"golang.org/x/crypto/ssh"
	"io"
)

type ClientChannel interface {
	io.ReadWriteCloser
	RequestChan() <-chan *ssh.Request

	SendRequest(string, bool, []byte) (bool, error) // TODO:Remove
}

type SSHClientChannel struct {
	ssh.Channel
	reqs <-chan *ssh.Request
}

func (c SSHClientChannel) RequestChan() <-chan *ssh.Request {
	return c.reqs
}

func NewSSHClientChannel(sc ssh.NewChannel) (ClientChannel, error) {
	cChannel, cReqs, err := sc.Accept()
	if err != nil {
		return nil, fmt.Errorf("Could not accept channel (%s)", err)
	}
	return SSHClientChannel{cChannel, cReqs}, nil
}

type WebSocketClientChannel struct {
	io.ReadWriteCloser
}

func NewWebSocketClientChannel(uuid string, ws *websocket.Conn) (ClientChannel, error) {
	c := WebSocketClientChannel{
		ReadWriteCloser: wsWrap{ws},
	}
	return c, nil
}

type wsWrap struct {
	core *websocket.Conn
}

func (w wsWrap) Close() error {
	return w.core.Close()
}
func (w wsWrap) Write(p []byte) (int, error) {
	return len(p), w.core.WriteMessage(websocket.TextMessage, p)
}
func (w wsWrap) Read(p []byte) (int, error) {
	_, bs, err := w.core.ReadMessage()
	copy(p, bs)
	return len(bs), err
}

func (WebSocketClientChannel) SendRequest(string, bool, []byte) (bool, error) { // TODO:Remove
	return false, fmt.Errorf("websocket client hasn't implement SendRequest")
}

func (WebSocketClientChannel) RequestChan() <-chan *ssh.Request {
	return nil
}
