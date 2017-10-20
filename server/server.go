package main

import (
	"fmt"
	"golang.org/x/crypto/ssh"
	"os"
	"time"
)

func main() {
	err := connectToHost("anything", "localhost:2200")
	if err != nil {
		fmt.Println("ERR:", err)
	}
}

func connectToHost(user, host string) error {
	sshConfig := &ssh.ClientConfig{
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}
	client, err := ssh.Dial("tcp", host, sshConfig)
	if err != nil {
		return err
	}

	_, uuid, err := client.SendRequest("hackme", true, nil)
	if err != nil {
		return err
	}
	fmt.Printf("UUID is %q\n", string(uuid))

	server, requests, err := client.OpenChannel("hackme", nil)
	if err != nil {
		return err
	}
	go makeChatRobot(server)

	r, w, _ := os.Pipe()
	go UIServer(r, "127.0.0.1:8080")

	return makeBashServer(server, requests, w)
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
