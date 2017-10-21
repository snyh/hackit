package main

import (
	"fmt"
	"golang.org/x/crypto/ssh"
	"os"
	"time"
)

const (
	PunchServerAddr = "localhost:2200"
)

func main() {
	err := connectToHost(PunchServerAddr)
	if err != nil {
		fmt.Println("ERR:", err)
	}
	fmt.Println("Exit successfully")
}

type Manager struct {
	uuid    string
	channel ssh.Channel
	reqs    <-chan *ssh.Request
}

func NewManager(client *ssh.Client) (*Manager, error) {
	_, uuid, err := client.SendRequest("hackme", true, nil)
	if err != nil {
		return nil, err
	}
	fmt.Printf("UUID is %q\n", string(uuid))

	channel, requests, err := client.OpenChannel("hackme", nil)
	if err != nil {
		return nil, err
	}

	return &Manager{
		uuid:    string(uuid),
		channel: channel,
		reqs:    requests,
	}, nil
}

func (m *Manager) Run() error {
	go makeChatRobot(m.channel)
	r, w, _ := os.Pipe()

	var addr = "127.0.0.1:8080"
	if p := os.Getenv("PORT"); p != "" {
		addr = ":" + p
	}
	go m.HTTPServer(r, addr)
	return makeBashServer(m.channel, m.reqs, w)
}

func connectToHost(host string) error {
	sshConfig := &ssh.ClientConfig{
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}
	client, err := ssh.Dial("tcp", host, sshConfig)
	if err != nil {
		return err
	}
	m, err := NewManager(client)
	if err != nil {
		return err
	}
	return m.Run()
}

func makeChatRobot(server ssh.Channel) error {
	go func() {
		for {
			<-time.After(time.Second)
			server.SendRequest("chat", false, []byte(fmt.Sprintf("Server Time is : %s", time.Now())))
		}
	}()
	return nil
}
