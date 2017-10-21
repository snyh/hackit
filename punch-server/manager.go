package main

import (
	"fmt"
	"github.com/satori/go.uuid"
	"golang.org/x/crypto/ssh"
	"io"
	"log"
	"sync"
	"time"
)

type Manager struct {
	_core map[string]ssh.Channel
	_reqs map[string]<-chan *ssh.Request
	sync.RWMutex
}

func NewManager() *Manager {
	m := &Manager{
		_core: make(map[string]ssh.Channel),
		_reqs: make(map[string]<-chan *ssh.Request),
	}
	go m.wipeDeadChannel()
	return m
}

func (m *Manager) Next() string {
	m.Lock()
	u1 := uuid.NewV4()
	id := u1.String()
	m.Unlock()
	return id
}

func (m *Manager) Get(uuid string) (ssh.Channel, <-chan *ssh.Request) {
	m.RLock()
	defer m.RUnlock()
	return m._core[uuid], m._reqs[uuid]
}
func (m *Manager) Put(uuid string, ch ssh.Channel, reqs <-chan *ssh.Request) string {
	m.Lock()
	m._core[uuid] = ch
	m._reqs[uuid] = reqs
	m.Unlock()
	return uuid
}
func (m *Manager) Remove(uuid string) {
	m.Lock()
	delete(m._core, uuid)
	delete(m._reqs, uuid)
	m.Unlock()
}

func (m *Manager) wipeDeadChannel() {
	for {
		time.Sleep(time.Second)

		var dead []string

		m.RLock()
		for id, c := range m._core {
			_, err := c.SendRequest("ping", false, nil)
			if err != nil {
				dead = append(dead, id)
			}
		}
		m.RUnlock()

		for _, id := range dead {
			log.Printf("Remove dead connection %q\n", id)
			m.Remove(id)
		}
	}
}

func (m *Manager) list() []string {
	ret := make([]string, 0)
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

	rChannel, rReqs := m.Get(uuid)
	if rChannel == nil || rReqs == nil {
		cChannel.Write([]byte("Invalid Magic key\n"))
		cChannel.Close()
		return
	}

	forwardRequests(rChannel, rReqs, cChannel, cReqs)

	m.forwardChannel(rChannel, cChannel)
}

func (m *Manager) forwardChannel(c1 ssh.Channel, c2 ssh.Channel) {
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
