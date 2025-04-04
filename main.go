package main

import (
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

type Server struct {
	Config
	peers map[*Peer]bool
	ln net.Listener
	addPeerCh chan *Peer
	quitCh chan struct{}
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
	}
}

func (s *Server) Start() error {
	ln, err := net.Listen("tcp", s.Port)
	if err != nil {
		return err
	}

	s.ln = ln

	go s.loop()

	slog.Info("Server started", "port", s.Port)

	return s.acceptLoop()
}

func (s *Server) loop() {
	for {
		select {
			case peer := <-s.addPeerCh:
				s.peers[peer] = true
			case <-s.quitCh:
				return
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
	peer := NewPeer(conn)
	s.addPeerCh <- peer
	slog.Info("new peer connected", "addr", conn.RemoteAddr())
	peer.readLoop()
}

func main() {
	server := NewServer(Config{})
	log.Fatal(server.Start())
}