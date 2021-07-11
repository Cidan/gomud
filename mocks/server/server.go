package server

import (
	"fmt"
	"net"
)

type Server struct {
	listener net.Listener
	Port     int
}

func New(port int) *Server {
	return &Server{
		Port: port,
	}
}

type connectFn func(net.Conn)

// Listen on a port for player connections.
func (s *Server) Listen(fn connectFn) error {
	l, err := net.Listen("tcp", fmt.Sprintf(":%d", s.Port))
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
	fmt.Printf("listener is %v\n", s.listener)
	s.listener.Close()
}
