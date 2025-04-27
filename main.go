package main

import (
	"flag"
	"fmt"
	"log"
	"log/slog"
	"net"
)

const (
	defaultPort = ":5001"
)

type Config struct {
	Port string
}

type Message struct {
	cmd Command
	peer *Peer
}
type Server struct {
	Config
	peers map[*Peer]bool
	ln net.Listener
	addPeerCh chan *Peer
	quitCh chan struct{}
	msgChan chan Message
	delPeerCh chan *Peer

	kv *KV
}

func NewServer(cfg Config) *Server {
	if len(cfg.Port) == 0 {
		cfg.Port = defaultPort
	}

	return &Server{
		Config: cfg,
		peers: make(map[*Peer]bool),
		addPeerCh: make(chan *Peer),
		quitCh: make(chan struct{}),
		msgChan: make(chan Message),
		delPeerCh: make(chan *Peer),
		kv: NewKV(),
	}
}

func (s *Server) Start() error {
	ln, err := net.Listen("tcp", s.Port)
	if err != nil {
		return err
	}

	s.ln = ln

	go s.loop()

	slog.Info("goredis server running", "port", s.Port)

	return s.acceptLoop()
}

func (s *Server) loop() {
	for {
		select {
			case peer := <-s.addPeerCh:
				slog.Info("peer connected", "remoteAddr", peer.conn.RemoteAddr())
				s.peers[peer] = true
			case <-s.quitCh:
				return
			case rawMsg := <-s.msgChan:
				if err := s.handleRawMsg(rawMsg); err != nil {
					slog.Error("handle raw message error", "error", err)
				}
			case peer := <- s.delPeerCh:
				slog.Info("peer disconnected", "remoteAddr", peer.conn.RemoteAddr())
				delete(s.peers, peer)
		}
	}
}

func (s *Server) acceptLoop() error {
	for {
		conn, err := s.ln.Accept()
		if err != nil {
			slog.Error("accept error", "error", err)
			continue
		}
		go s.handleConn(conn)
	}
}

func (s *Server) handleConn(conn net.Conn) {
	peer := NewPeer(conn, s.msgChan, s.delPeerCh)
	s.addPeerCh <- peer
	if err := peer.readLoop(); err != nil {
		slog.Error("peer read error", "err", err, "remote_addr", conn.RemoteAddr())
	}
}

func (s *Server) handleRawMsg(msg Message) error {
	switch v := msg.cmd.(type) {
	case SetCommand:
		return s.kv.Set(v.key, v.val)
	case GetCommand:
		val, ok := s.kv.Get(v.key)
		if !ok {
			return fmt.Errorf("key not found")
		}
		_, err := msg.peer.Send(val)
		if err != nil {
			slog.Error("peer send error", "err", err)
		}
	}

	return nil
}

func main() {
	listenAddr := flag.String("listenAddr", defaultPort, "listen address of the goredis server")
	flag.Parse()
	server := NewServer(Config{
		Port: *listenAddr,
	})
	log.Fatal(server.Start())
}