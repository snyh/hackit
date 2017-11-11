package main

import (
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/kr/pty"
	"golang.org/x/crypto/ssh"
	"io"
	"log"
	"net/http"
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

	chatQueue ChatQueue

	channel       ssh.Channel
	inSSHRequests <-chan *ssh.Request

	shell *os.File //bash的输入输出, 实际由pty包装

	observer WriteSwitcher
	once     sync.Once
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

		channel:       channel,
		inSSHRequests: requests,

		observer: NewSimpleWriteSwitcher(),
	}, nil
}

func (c *HackItConn) SetShellObserver(p io.Writer) { c.observer.Switch(p) }

// ServerTTY 打印HackItConn的内容到本地ws中，以便被控者可以看到操控者执行的具体命令
func (c *HackItConn) ServeTTY(w http.ResponseWriter, r *http.Request) {
	rp, wp, err := os.Pipe()
	if err != nil {
		writeJSON(w, 501, err)
		return
	}
	c.SetShellObserver(wp)

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
		io.Copy(wsWrap{ws}, rp)
		done <- struct{}{}
		ws.Close()
	}()
	go wsPing(ws, done)
}

func (c *HackItConn) handleInSSHRequest() {
	for req := range c.inSSHRequests {
		switch req.Type {
		case "chat":
			c.onReceiveChat(req.Payload)
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
		io.Copy(c.channel, io.TeeReader(c.shell, c.observer))
		c.Stop()
	}()
	go func() {
		io.Copy(c.shell, c.channel)
		c.Stop()
	}()
	return nil
}
