package main

import (
	"fmt"
	"github.com/kr/pty"
	"golang.org/x/crypto/ssh"
	"io"
	"log"
	"os"
	"os/exec"
	"sync"
	"time"
)

type HackItConnStatus string

type HackItConn struct {
	UUID     string
	Status   HackItConnStatus
	CreateAt time.Time

	chatBuffer *ChatBuffer

	channel       *ChannelHistory
	inSSHRequests <-chan *ssh.Request

	shell *os.File //bash的输入输出, 实际由pty包装

	once sync.Once
}

func NewHackItConn(serveAddr string) (*HackItConn, error) {
	sshConfig := &ssh.ClientConfig{
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}
	client, err := ssh.Dial("tcp", serveAddr, sshConfig)
	if err != nil {
		return nil, err
	}
	_, uuid, err := client.SendRequest("hackme", true, nil)
	if err != nil {
		return nil, err
	}
	channel, requests, err := client.OpenChannel("hackme", nil)
	if err != nil {
		return nil, err
	}

	return &HackItConn{
		UUID: string(uuid),

		Status:   "ready",
		CreateAt: time.Now().UTC(),

		channel:       NewChannelHistory(channel),
		inSSHRequests: requests,

		chatBuffer: NewChatBuffer(channel),
	}, nil
}

func (c *HackItConn) handleInSSHRequest() {
	for req := range c.inSSHRequests {
		switch req.Type {
		case "chat":
			c.chatBuffer.WriteFromSSH(req.Payload)
			req.Reply(true, nil)
		case "shell", "ping":
			req.Reply(true, nil)
		case "pty-req":
			termLen := req.Payload[3]
			w, h := parseDims(req.Payload[termLen+4:])
			log.Print("Creating pty...", w, h)
			SetWinsize(c.shell.Fd(), w, h)
			req.Reply(true, nil)
		case "window-change":
			w, h := parseDims(req.Payload)
			SetWinsize(c.shell.Fd(), w, h)
		case "hacking":
			c.Status = "running"
			req.Reply(true, nil)
		default:
			fmt.Println("bad things..", req.Type)
			if req.WantReply {
				req.Reply(true, nil)
			}
		}
	}
	fmt.Println("EndOfRequest...")
}

func (c *HackItConn) Stop() error {
	close := func() {
		c.Status = "closed"
		c.channel.Close()
		c.shell.Close()
		log.Printf("Session closed")
	}
	c.once.Do(close)
	return nil
}

func (c *HackItConn) Start() error {
	if c.shell != nil {
		panic("can't running again")
	}

	var err error

	bash := exec.Command("bash")
	c.shell, err = pty.Start(bash)
	if err != nil {
		_, err := bash.Process.Wait()
		if err != nil {
			log.Printf("Failed to exit bash (%s)", err)
		}
		c.Stop()
		return err
	}

	c.Status = "ready"

	go c.handleInSSHRequest()
	go func() {
		io.Copy(c.channel, c.shell)
		c.Stop()
	}()
	go func() {
		io.Copy(c.shell, c.channel)
		c.Stop()
	}()
	return nil
}
