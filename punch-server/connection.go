package main

import (
	"github.com/gorilla/websocket"
	"golang.org/x/crypto/ssh"
	"io"
	"log"
	"sync"
)

type HackerConn struct {
	SSHChannel    ssh.Channel
	InSSHRequests <-chan *ssh.Request

	ClientChannel io.ReadWriteCloser

	chatBuffer *ChatBuffer
	closeOnce  sync.Once
}

func NewHackerConn(channel ssh.Channel, reqs <-chan *ssh.Request) *HackerConn {
	return &HackerConn{
		SSHChannel:    channel,
		InSSHRequests: reqs,
		chatBuffer:    NewChatBuffer(channel),
	}
}

func (c *HackerConn) Start(ws *websocket.Conn) {
	c.ClientChannel = wsWrap{ws}

	c.SSHChannel.SendRequest("hacking", false, nil)

	makeChatRobot(c.SSHChannel)

	c.forwardRequests()
	c.forwardChannel()
}

func (c *HackerConn) Close() error {
	close := func() {
		c.SSHChannel.Close()
		if c.ClientChannel != nil {
			c.ClientChannel.Close()
		}
	}
	c.closeOnce.Do(close)
	return nil
}

func (c *HackerConn) forwardRequests() {
	go func() {
		for req := range c.InSSHRequests {
			if req.WantReply {
				req.Reply(true, nil)
			}
			log.Println("F　cR ---> sC", string(req.Payload))
			//c.ClientChannel.SendRequest(req.Type, req.WantReply, req.Payload)
		}
	}()
}

func (c *HackerConn) forwardChannel() {
	go func() {
		io.Copy(c.SSHChannel, c.ClientChannel)
		c.Close()
	}()

	io.Copy(c.ClientChannel, c.SSHChannel)
	c.Close()
}
