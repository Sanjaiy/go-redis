package main

import (
	"fmt"
	"net"
)

type Peer struct {
	conn net.Conn
}

func NewPeer(conn net.Conn) *Peer {
	return &Peer{
		conn: conn,
	}
}

func (p *Peer) readLoop() {
	buf := make([]byte, 1024)

	for {
				
	}
}