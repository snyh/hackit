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

	channel ssh.Channel
	reqs    <-chan *ssh.Request
	tty     *os.File

	printer WriteSwitcher
	once    sync.Once
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
		UUID:     string(uuid),
		Status:   "ready",
		CreateAt: time.Now().UTC(),

		channel: channel,
		reqs:    requests,
		printer: NewSimpleSwitcher(),
	}, nil
}

func (c *HackItConn) AttachPrinter(p io.Writer) { c.printer.Switch(p) }

func (c *HackItConn) handleRequest() {
	for req := range c.reqs {
		switch req.Type {
		case "chat":
			//			fmt.Printf("\033[31m %s \033[0m\n", req.Payload)
			req.Reply(true, nil)
		case "shell", "ping":
			req.Reply(true, nil)
		case "pty-req":
			termLen := req.Payload[3]
			w, h := parseDims(req.Payload[termLen+4:])
			log.Print("Creating pty...", w, h)
			SetWinsize(c.tty.Fd(), w, h)
			req.Reply(true, nil)
		case "window-change":
			w, h := parseDims(req.Payload)
			SetWinsize(c.tty.Fd(), w, h)
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
		c.tty.Close()
		log.Printf("Session closed")
	}
	c.once.Do(close)
	return nil
}

func (c *HackItConn) Start() error {
	if c.tty != nil {
		panic("can't running again")
	}

	var err error

	bash := exec.Command("bash")
	c.tty, err = pty.Start(bash)
	if err != nil {
		_, err := bash.Process.Wait()
		if err != nil {
			log.Printf("Failed to exit bash (%s)", err)
		}
		c.Stop()
		return err
	}

	c.Status = "running"

	go c.handleRequest()
	go func() {
		io.Copy(c.channel, io.TeeReader(c.tty, c.printer))
		c.Stop()
	}()
	go func() {
		io.Copy(c.tty, c.channel)
		c.Stop()
	}()
	return nil
}
