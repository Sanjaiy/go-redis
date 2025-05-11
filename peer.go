package main

import (
	"fmt"
	"io"
	"log"
	"net"

	"github.com/tidwall/resp"
)

type Peer struct {
	conn net.Conn
	msgChan chan Message
	delCh chan *Peer
}

func NewPeer(conn net.Conn, msgChan chan Message, delCh  chan *Peer) *Peer {
	return &Peer{
		conn: conn,
		msgChan: msgChan,
		delCh: delCh,
	}
}

func (p *Peer) readLoop() error {
	rd := resp.NewReader(p.conn)

	for  {
		v, _, err := rd.ReadValue()
		if err == io.EOF {
			p.delCh <- p
			break
		}
		if err != nil {
			log.Fatal(err)
		}

		var cmd Command
		if v.Type() == resp.Array {
			rawCMD := v.Array()[0]
			switch rawCMD.String() {
				case CommandSET:
					if len(v.Array()) != 3 {
						return fmt.Errorf("invalid number of variables for SET command")
					}
					cmd = SetCommand{
						key: v.Array()[1].Bytes(),
						val: v.Array()[2].Bytes(),
					}
				case CommandGET:
					if len(v.Array()) != 2 {
						return fmt.Errorf("invalid number of variables for GET command")
					}
					cmd = GetCommand{
						key: v.Array()[1].Bytes(),
					}
				case CommandHELLO:
					cmd = HelloCommand{
						value: v.Array()[1].String(),
					}
				case CommandClient:
					cmd = ClientCommand{
						value: v.Array()[1].String(),
					}
				default:
					fmt.Printf("got unknown command => %+v\n", rawCMD)
			}
			p.msgChan <- Message{
				cmd: cmd,
				peer: p,
			}
		}
	}
	
	return nil
}

func (p *Peer) Send(msg []byte) (int, error) {
	return p.conn.Write(msg)
}