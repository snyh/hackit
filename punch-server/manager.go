package main

import (
	"fmt"
	"github.com/satori/go.uuid"
	"golang.org/x/crypto/ssh"
	"log"
	"sync"
	"time"
)

type Manager struct {
	sync.RWMutex

	conns map[string]*HackerConn
}

func NewManager() *Manager {
	m := &Manager{
		conns: make(map[string]*HackerConn),
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

func (m *Manager) FindConnection(uuid string) *HackerConn {
	m.RLock()
	defer m.RUnlock()
	return m.conns[uuid]
}

func (m *Manager) PutConnection(uuid string, conn *HackerConn) string {
	m.Lock()
	m.conns[uuid] = conn
	m.Unlock()
	return uuid
}

func (m *Manager) Remove(uuid string) {
	m.Lock()
	c := m.conns[uuid]
	delete(m.conns, uuid)
	m.Unlock()

	if c != nil {
		c.Close()
	}
}

func (m *Manager) wipeDeadChannel() {
	for {
		time.Sleep(time.Second)

		var dead []string

		m.RLock()
		for id, c := range m.conns {
			_, err := c.SSHChannel.SendRequest("ping", false, nil)
			if err != nil {
				dead = append(dead, id)
			}
		}
		m.RUnlock()

		for _, id := range dead {
			log.Printf("Remove dirty connection %q\n", id)
			m.Remove(id)
		}
	}
}

func (m *Manager) list() []string {
	ret := make([]string, 0)
	for id := range m.conns {
		ret = append(ret, id)
	}
	return ret
}

func makeChatRobot(server ssh.Channel) error {
	go func() {
		for {
			<-time.After(time.Second * 3)

			msg := fmt.Sprintf("Server Time is : %s", time.Now())
			str := fmt.Sprintf(`{"author": "them",
"type": "text",
"data": {
   "text": %q
}
}`, msg)
			server.SendRequest("chat", false, []byte(str))
		}
	}()
	return nil
}
