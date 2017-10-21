package main

import (
	"encoding/binary"
	"fmt"
	"github.com/kr/pty"
	"golang.org/x/crypto/ssh"
	"io"
	"log"
	"os"
	"os/exec"
	"sync"
	"syscall"
	"unsafe"
)

func handleRequest(bashf *os.File, reqs <-chan *ssh.Request) {
	for req := range reqs {
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
			SetWinsize(bashf.Fd(), w, h)
			req.Reply(true, nil)
		case "window-change":
			w, h := parseDims(req.Payload)
			SetWinsize(bashf.Fd(), w, h)

		default:
			fmt.Println("bad things..", req.Type)
			if req.WantReply {
				req.Reply(true, nil)
			}

		}
	}
	fmt.Println("EndOfRequest...")
}

func makeBashServer(connection ssh.Channel, reqs <-chan *ssh.Request, printer io.Writer) error {
	// Fire up bash for this session
	bash := exec.Command("bash")

	// Prepare teardown function
	close := func() {
		connection.Close()
		_, err := bash.Process.Wait()
		if err != nil {
			log.Printf("Failed to exit bash (%s)", err)
		}
		log.Printf("Session closed")
	}

	// Allocate a terminal for this channel
	bashf, err := pty.Start(bash)

	if err != nil {
		close()
		return err
	}
	//pipe session to bash and visa-versa
	var once sync.Once
	go func() {
		io.Copy(connection, io.TeeReader(bashf, printer))
		once.Do(close)
	}()
	go func() {
		io.Copy(bashf, connection)
		once.Do(close)
	}()

	handleRequest(bashf, reqs)
	return nil
}

// =======================

// parseDims extracts terminal dimensions (width x height) from the provided buffer.
func parseDims(b []byte) (uint32, uint32) {
	w := binary.BigEndian.Uint32(b)
	h := binary.BigEndian.Uint32(b[4:])
	return w, h
}

// ======================

// Winsize stores the Height and Width of a terminal.
type Winsize struct {
	Height uint16
	Width  uint16
	x      uint16 // unused
	y      uint16 // unused
}

// SetWinsize sets the size of the given pty.
func SetWinsize(fd uintptr, w, h uint32) {
	ws := &Winsize{Width: uint16(w), Height: uint16(h)}
	syscall.Syscall(syscall.SYS_IOCTL, fd, uintptr(syscall.TIOCSWINSZ), uintptr(unsafe.Pointer(ws)))
}
