package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

var (
	PunchServerAddr = "localhost:2200"
	LocalServerAddr = "localhost:8080"
)

func main() {
	if p := os.Getenv("PORT"); p != "" {
		LocalServerAddr = ":" + p
	}
	m, err := NewManager(PunchServerAddr, LocalServerAddr)
	if err != nil {
		fmt.Println("ERR:", err)
	}
	if err := m.Run(); err != nil {
		fmt.Println("ERR:", err)
	}
}

type Status string

const (
	StatusOnline    = "online"
	StatusListen    = "listen"
	StatusConnected = "connected"
	StatusError     = "error"
)

type Manager struct {
	status     Status
	hackitAddr string
	localAddr  string
}

func NewManager(hackitAddr string, localAddr string) (*Manager, error) {
	m := &Manager{
		status:     StatusOnline,
		hackitAddr: hackitAddr,
		localAddr:  localAddr,
	}
	return m, nil
}

func (m *Manager) openHackIt(apiServer string) (*HackItConn, error) {
	return NewHackItConn(apiServer)
}

func (m *Manager) Run() error {
	conn, err := NewHackItConn(m.hackitAddr)
	if err != nil {
		return err
	}
	r, w, err := os.Pipe()
	if err != nil {
		return err
	}
	go HTTPServer(r, m.localAddr, conn.uuid)
	return conn.Run(w)
}

func HTTPServer(f io.Reader, addr string, uuid string) {
	http.HandleFunc("/tty/status", serveStatus(uuid))
	http.HandleFunc("/tty", serveWs(f))
	go func() {
		time.Sleep(time.Millisecond * 20)
		openUrl("http://" + addr)
	}()
	log.Fatal(http.ListenAndServe(addr, nil))
}

func serveStatus(uuid string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fixCSR(w)
		writeJSON(w, uuid)
	}
}
