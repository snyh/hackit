package main

import (
	"fmt"
	"github.com/satori/go.uuid"
	"golang.org/x/crypto/ssh"
	"io"
	"sync"
)

type Manager struct {
	_core map[string]ssh.Channel
	_reqs map[string]<-chan *ssh.Request
}

func NewManager() *Manager {
	return &Manager{
		_core: make(map[string]ssh.Channel),
		_reqs: make(map[string]<-chan *ssh.Request),
	}
}

func (m *Manager) Match(uuid string) (ssh.Channel, <-chan *ssh.Request) {
	return m._core[uuid], m._reqs[uuid]
}
func (m *Manager) Next() string {
	u1 := uuid.NewV4()
	id := u1.String()
	id = "12345"
	return id
}
func (m *Manager) Put(id string, ch ssh.Channel, reqs <-chan *ssh.Request) string {
	fmt.Println("PUT>>>>", id)
	m._core[id] = ch
	m._reqs[id] = reqs
	return id
}

func (m *Manager) list() []string {
	var ret []string
	for id := range m._core {
		ret = append(ret, id)
	}
	return ret
}

func (m *Manager) Hacking(newChannel ssh.NewChannel, uuid string) {
	cChannel, cReqs, err := newChannel.Accept()
	if err != nil {
		fmt.Printf("Could not accept channel (%s)", err)
		return
	}
	rChannel, rReqs := m.Match(uuid)

	forwardRequests(rChannel, rReqs, cChannel, cReqs)
	forwardChannel(rChannel, cChannel)
}

func forwardChannel(c1 ssh.Channel, c2 ssh.Channel) {
	close := func() {
		c1.Close()
		c2.Close()
	}
	one := sync.Once{}
	go func() {
		io.Copy(c1, c2)
		one.Do(close)
	}()
	go func() {
		io.Copy(c2, c1)
		one.Do(close)
	}()
}

func forwardRequests(cC ssh.Channel, cR <-chan *ssh.Request, sC ssh.Channel, sR <-chan *ssh.Request) {
	go func() {
		for req := range cR {
			if req.WantReply {
				req.Reply(true, nil)
			}
			//			fmt.Println("F　cR ---> sC", string(req.Payload))
			sC.SendRequest(req.Type, req.WantReply, req.Payload)
		}
	}()

	go func() {
		for req := range sR {
			if req.WantReply {
				req.Reply(true, nil)
			}
			//			fmt.Println("F sR ---> cC", string(req.Payload))
			cC.SendRequest(req.Type, req.WantReply, req.Payload)
		}
	}()
}
