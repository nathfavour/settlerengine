package uds

import (
	"log"
	"net"
	"os"
)

type Server struct {
	path string
	l    net.Listener
}

func NewServer(path string) *Server {
	return &Server{path: path}
}

func (s *Server) Start() error {
	// Remove existing socket if any
	if err := os.RemoveAll(s.path); err != nil {
		return err
	}

	l, err := net.Listen("unix", s.path)
	if err != nil {
		return err
	}
	s.l = l

	log.Printf("🔌 UDS Server: Listening on %s", s.path)

	go s.accept()
	return nil
}

func (s *Server) accept() {
	for {
		conn, err := s.l.Accept()
		if err != nil {
			return
		}
		go s.handleConn(conn)
	}
}

func (s *Server) handleConn(conn net.Conn) {
	defer conn.Close()
	// Placeholder for UDS protocol handling
	// This could allow local agents to request signatures, check payment status, etc.
	log.Printf("📡 UDS: New local connection from %s", conn.RemoteAddr())
	
	// Example: just echo or close for now until protocol is defined
}

func (s *Server) Close() error {
	if s.l != nil {
		return s.l.Close()
	}
	return nil
}
