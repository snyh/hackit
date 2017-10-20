package main

import (
	"fmt"
	"golang.org/x/crypto/ssh"
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

	s, err := client.NewSession()
	s.Stdin = os.Stdin
	s.Stdout = os.Stdout
	s.RequestPty("xterm", 40, 80, ssh.TerminalModes{
		ssh.ECHO:  0, // Disable echoing
		ssh.IGNCR: 1, // Ignore CR on input.
	})
	s.Shell()
	return s.Wait()

	// conn, chats, err := client.OpenChannel("hacking", nil)
	// RequestPty(conn, "xterm", 40, 80, ssh.TerminalModes{
	// 	ssh.ECHO:  0, // Disable echoing
	// 	ssh.IGNCR: 1, // Ignore CR on input.
	// })

	// go handleRequest(chats)
	// //	go makeChatRobot(conn)
	// go func() {
	// 	io.Copy(os.Stdout, conn)
	// }()
	// io.Copy(conn, os.Stdin)
}

func RequestPty(channel ssh.Channel, term string, h, w int, termmodes ssh.TerminalModes) error {
	var tm []byte
	for k, v := range termmodes {
		kv := struct {
			Key byte
			Val uint32
		}{k, v}

		tm = append(tm, ssh.Marshal(&kv)...)
	}
	tm = append(tm, 0)
	// RFC 4254 Section 6.2.
	req := struct {
		Term     string
		Columns  uint32
		Rows     uint32
		Width    uint32
		Height   uint32
		Modelist string
	}{
		Term:     term,
		Columns:  uint32(w),
		Rows:     uint32(h),
		Width:    uint32(w * 8),
		Height:   uint32(h * 8),
		Modelist: string(tm),
	}
	ok, err := channel.SendRequest("pty-req", true, ssh.Marshal(&req))
	if err == nil && !ok {
		err = fmt.Errorf("pyt-req failed: %v", err)
	}
	return err
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
