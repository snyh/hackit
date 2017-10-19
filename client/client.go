package main

import (
	"fmt"
	"golang.org/x/crypto/ssh"
	"io"
	"os"
	"time"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Specify magic code")
		return
	}
	uuid := os.Args[1]

	err := connectToHost(uuid, "localhost:2200")
	if err != nil {
		fmt.Println("ERR:", err)
	}
}

func connectToHost(uuid, host string) error {
	sshConfig := &ssh.ClientConfig{
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}
	client, err := ssh.Dial("tcp", host, sshConfig)
	if err != nil {
		return err
	}
	client.SendRequest("hacking", true, []byte(uuid))

	conn, chats, err := client.OpenChannel("hacking", nil)
	go handleRequest(chats)

	go makeChatRobot(conn)
	return login(conn)
}

func makeChatRobot(server ssh.Channel) error {
	go func() {
		for {
			<-time.After(time.Second)
			server.SendRequest("chat", false, []byte(fmt.Sprintf("Client Time is : %s", time.Now())))
		}
	}()
	return nil
}

func handleRequest(reqs <-chan *ssh.Request) {
	for req := range reqs {
		switch req.Type {
		case "chat":
			//			fmt.Printf("\033[31m %s \033[0m\n", req.Payload)
			req.Reply(true, nil)
		default:
			fmt.Println("bad things..", req.Type)
			req.Reply(false, nil)
		}
	}
	fmt.Println("EndOfRequest...")
}

func login(conn ssh.Channel) error {
	go func() {
		io.Copy(conn, os.Stdin)
	}()
	io.Copy(os.Stdout, conn)
	return nil
}
