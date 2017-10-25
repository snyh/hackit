package main

import (
	"flag"
	"fmt"
	"golang.org/x/crypto/ssh"
	"io/ioutil"
	"log"
	"net"
	"os"
)

type Config struct {
	SSH  string
	UI   string
	CERT string
}

func parse() Config {
	var sshAddr, uiAddr, cert string
	flag.StringVar(&sshAddr, "ssh_listen", ":2200", "the ssh channel listen address")
	flag.StringVar(&uiAddr, "http_listen", ":80", "the http ui server listen address")
	flag.StringVar(&cert, "cert", "./cert", "the path of certificate file, generate by ssh-keygen -t rsa")
	flag.Parse()
	return Config{
		sshAddr, uiAddr, cert,
	}
}

func buildSSHConfig(cert string) (*ssh.ServerConfig, error) {
	config := &ssh.ServerConfig{
		NoClientAuth: true,
	}
	// You can generate a keypair with 'ssh-keygen -t rsa'
	privateBytes, err := ioutil.ReadFile(cert)
	if err != nil {
		return nil, err
	}

	private, err := ssh.ParsePrivateKey(privateBytes)
	if err != nil {
		return nil, err
	}
	config.AddHostKey(private)
	return config, nil
}

func main() {
	c := parse()

	sshConfig, err := buildSSHConfig(c.CERT)
	if err != nil {
		log.Fatalf("Can't setup ssh protocol %s", err)
		return
	}

	listener, err := net.Listen("tcp", c.SSH)
	if err != nil {
		log.Fatalf("Failed to listen %s.", err)
		return
	}
	log.Printf("Listening ssh on %s\n", listener.Addr())
	m := NewManager()

	go func() {
		var err = UIServer(c.SSH, c.UI, m)
		if err != nil {
			log.Fatalf("Can't setup http protocol %s", err)
			os.Exit(-1)
		}
	}()

	for {
		tcpConn, err := listener.Accept()
		if err != nil {
			log.Printf("Failed to accept incoming connection (%s)", err)
			continue
		}
		fmt.Println("LLLLLLLLLL:", tcpConn.LocalAddr())
		go dispatch(m, tcpConn, sshConfig)
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
