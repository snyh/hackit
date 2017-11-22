package main

import (
	"flag"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"net"
	"net/http"
	"os"
	"time"
)

func OpenBrowser(uiServer string, port string) {
	if DEV {
		return
	}
	go func() {
		time.Sleep(time.Millisecond * 500)
		openUrl(fmt.Sprintf("%s/mysys/%s", uiServer, port))
	}()
}

const DEV = true

func main() {
	var remoteHTTPURL, localAddr, apiAddr string

	if DEV {
		flag.StringVar(&remoteHTTPURL, "remote", "http://localhost:8080", "the hackit's http address.")
		flag.StringVar(&apiAddr, "api", "localhost:2200", "the hackit's api address")
	} else {
		flag.StringVar(&remoteHTTPURL, "remote", "http://hackit.snyh.org", "the server address")
		flag.StringVar(&apiAddr, "api", "hackit.snyh.org:2200", "the hackit's api address")
	}

	flag.StringVar(&localAddr, "local", "auto", "the local listen address")

	flag.Parse()

	if p := os.Getenv("PORT"); p != "" {
		localAddr = ":" + p
	}

	m, err := NewManager(apiAddr, localAddr)
	if err != nil {
		fmt.Println("ERR:", err)
	}
	OpenBrowser(remoteHTTPURL, m.port)

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

	conns map[string]*HackItConn

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
	r.HandleFunc("/listTTYs", m.handleListConns)
	r.HandleFunc("/tty/{uuid:[a-z0-9-]+}", m.ServeTTY)
	r.HandleFunc("/chat/{uuid:[a-z0-9-]+}", m.ServeChat)
	r.HandleFunc("/requestTTY", m.handleNewConnect)

	http.Handle("/", r)
	return http.Serve(m.listener, nil)
}

func (m *Manager) handleNewConnect(w http.ResponseWriter, r *http.Request) {
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

// ServerTTY 打印HackItConn的内容到本地ws中，以便被控者可以看到操控者执行的具体命令
func (m *Manager) ServeTTY(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	conn, ok := m.conns[vars["uuid"]]
	if !ok {
		writeJSON(w, 404, "invalid magic key")
		return
	}
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}

	// setup websocket
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		writeJSON(w, 501, err)
		return
	}
	conn.channel.Switch(ws)
}

// ServeChat 收发WebSocket上的chat message 到c.chatQueue上
func (m *Manager) ServeChat(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	conn, ok := m.conns[vars["uuid"]]
	if !ok {
		writeJSON(w, 404, "invalid magic key")
		return
	}

	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}

	// setup websocket
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		writeJSON(w, 501, err)
		return
	}
	conn.chatBuffer.SwitchWS(ws)
}

func (m *Manager) handleStatus(w http.ResponseWriter, r *http.Request) {
	fixCSR(w)
	writeJSON(w, 200, "online")
}

func (m *Manager) handleListConns(w http.ResponseWriter, r *http.Request) {
	fixCSR(w)
	var ret = make([]*HackItConn, 0)
	for _, v := range m.conns {
		ret = append(ret, v)
	}
	writeJSON(w, 200, ret)
}
