package server

import (
	"fmt"
	"net"
)

type Server struct {
	listener net.Listener
}

func New() *Server {
	return &Server{}
}

type connectFn func(net.Conn)

// Listen on a port for player connections.
func (s *Server) Listen(port int, fn connectFn) error {
	l, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return err
	}
	s.listener = l

	// Loop for new connections
	for {
		c, err := l.Accept()
		if err != nil {
			break
		} else {
			go fn(c)
			continue
		}
	}
	return nil
}

func (s *Server) Close() {
	s.listener.Close()
}
