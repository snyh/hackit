package main

import (
	"fmt"
	"io"
	"net"
)

type PairLine struct {
	uuid string
	in   net.Conn
	out  net.Conn
	stop chan bool
}

func (l *PairLine) AttachIn(c net.Conn)  { l.in = c }
func (l *PairLine) AttachOut(c net.Conn) { l.out = c }
func (*PairLine) Stop() error            { return nil }
func (l *PairLine) Start() error {
	p1, p2 := net.Pipe()
	l.stop = make(chan bool)
	go func() {
		_, err := io.Copy(p1, l.out)
		if err != nil {
			fmt.Println(err)
		}
		l.stop <- true
	}()
	go func() {
		_, err := io.Copy(p2, l.in)
		if err != nil {
			fmt.Println(err)
		}
		l.stop <- true
	}()
	return nil
}
