package main

import (
	"flag"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"io"
	"net"
	"net/http"
	"os"
	"time"
)

func OpenBrowser(uiServer string, port string) {
	go func() {
		time.Sleep(time.Millisecond * 500)
		openUrl(fmt.Sprintf("%s/mysys/%s", uiServer, port))
	}()
}

func findAPIAddress(remote string) string {
	return "hackit.snyh.org:2200"
}

func main() {
	var remoteURL, localAddr string
	flag.StringVar(&remoteURL, "remote", "http://hackit.snyh.org", "the server address")
	flag.StringVar(&localAddr, "local", "auto", "the local listen address")
	flag.Parse()

	if p := os.Getenv("PORT"); p != "" {
		localAddr = ":" + p
	}

	m, err := NewManager(findAPIAddress(remoteURL), localAddr)
	if err != nil {
		fmt.Println("ERR:", err)
	}
	OpenBrowser(remoteURL, m.port)

	if err := m.Run(); err != nil {
		fmt.Println("ERR:", err)
	}
	fmt.Println("Exit successfully")
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
	conns      map[string]*HackItConn

	listener net.Listener
	port     string
}

func NewManager(hackitAddr string, localAddr string) (*Manager, error) {
	if localAddr == "auto" {
		localAddr = "127.0.0.1:0"
	}
	l, err := net.Listen("tcp", localAddr)
	if err != nil {
		return nil, err
	}
	_, p, err := net.SplitHostPort(l.Addr().String())
	if err != nil {
		return nil, err
	}

	m := &Manager{
		status:     StatusOnline,
		hackitAddr: hackitAddr,
		conns:      make(map[string]*HackItConn),
		listener:   l,
		port:       p,
	}
	return m, nil
}

func (m *Manager) openHackIt(apiServer string) (*HackItConn, error) {
	return NewHackItConn(apiServer)
}

func (m *Manager) Run() error {
	r := mux.NewRouter()
	r.HandleFunc("/status", m.handleStatus)
	r.HandleFunc("/listTTYs", m.handleListTTYs)
	r.HandleFunc("/tty/{uuid:[a-z0-9-]+}", m.handleConnect)
	r.HandleFunc("/requestTTY", m.handleRequestConnect)
	http.Handle("/", r)
	return http.Serve(m.listener, nil)
}

// TODO: remove this from global scope
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func (m *Manager) handleConnect(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	conn, ok := m.conns[vars["uuid"]]
	if !ok {
		writeJSON(w, 404, "invalid magic key")
		return
	}
	rp, wp, err := os.Pipe()
	if err != nil {
		writeJSON(w, 501, err)
		return
	}
	conn.AttachPrinter(wp)

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

func (m *Manager) handleRequestConnect(w http.ResponseWriter, r *http.Request) {
	fixCSR(w)
	conn, err := NewHackItConn(m.hackitAddr)
	if err != nil {
		writeJSON(w, 502, err.Error())
		return
	}
	err = conn.Start()
	if err != nil {
		writeJSON(w, 502, err.Error())
		return
	}

	m.conns[conn.UUID] = conn
	writeJSON(w, 200, conn.UUID)
}

func (m *Manager) handleStatus(w http.ResponseWriter, r *http.Request) {
	fixCSR(w)
	writeJSON(w, 200, "online")
}
func (m *Manager) handleListTTYs(w http.ResponseWriter, r *http.Request) {
	fixCSR(w)
	var ret = make([]*HackItConn, 0)
	for _, v := range m.conns {
		ret = append(ret, v)
	}
	writeJSON(w, 200, ret)
}
