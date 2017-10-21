package main

import (
	"golang.org/x/crypto/ssh"
	"io/ioutil"
	"log"
	"net"
)

const (
	SSHAddr = "0.0.0.0:2200"
	UIAddr  = ":2207"
)

func main() {
	config := &ssh.ServerConfig{
		NoClientAuth: true,
	}

	// You can generate a keypair with 'ssh-keygen -t rsa'
	privateBytes, err := ioutil.ReadFile("out")
	if err != nil {
		log.Fatal("Failed to load private key (./out)")
	}

	private, err := ssh.ParsePrivateKey(privateBytes)
	if err != nil {
		log.Fatal("Failed to parse private key")
	}

	config.AddHostKey(private)

	// Once a ServerConfig has been configured, connections can be accepted.
	listener, err := net.Listen("tcp", SSHAddr)
	if err != nil {
		log.Fatalf("Failed to listen on 2200 (%s)", err)
	}

	m := NewManager()

	go UIServer(UIAddr, m)

	// Accept all connections
	log.Printf("Listening on %s\n", SSHAddr)
	for {
		tcpConn, err := listener.Accept()
		if err != nil {
			log.Printf("Failed to accept incoming connection (%s)", err)
			continue
		}
		go dispatch(m, tcpConn, config)
	}
}

func dispatch(m *Manager, tcpConn net.Conn, config *ssh.ServerConfig) error {
	// Before use, a handshake must be performed on the incoming net.Conn.
	sshConn, channels, reqs, err := ssh.NewServerConn(tcpConn, config)
	if err != nil {
		log.Printf("Failed to handshake (%s)", err)
		return err
	}

	for req := range reqs {
		switch req.Type {
		case "hackme":
			id := m.Next()
			req.Reply(true, []byte(id))
			log.Printf("New SSH connection \033[31mHackMe\033[0m %s (%s)", sshConn.RemoteAddr(), sshConn.ClientVersion())

			// We only support one channel per ssh connection.
			nch := <-channels
			ch, reqs, err := nch.Accept()
			if err != nil {
				return err
			}
			m.Put(id, ch, reqs)
		case "hacking":
			log.Printf("New SSH connection \033[31mHacking\033[0m %s (%s)", sshConn.RemoteAddr(), sshConn.ClientVersion())
			req.Reply(true, nil)
			uuid := string(req.Payload)

			// We only support one channel per ssh connection.
			c, err := NewSSHClientChannel(<-channels)
			if err != nil {
				return err
			}
			go m.Hacking(c, uuid)
		default:
			return tcpConn.Close()
		}
	}
	return nil
}
