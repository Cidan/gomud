package server

import (
	"errors"
	"net"

	"github.com/Cidan/gomud/util"

	"github.com/rs/zerolog/log"
)

type ConnectionHandler func(net.Conn)

// Server is a network server construct
// that handles incoming player connections.
type Server struct {
	ln               net.Listener
	handleConnection ConnectionHandler
}

// New Server
func New() *Server {
	return &Server{}
}

func (s *Server) SetHandler(fn ConnectionHandler) {
	s.handleConnection = fn
}

// Listen on a port for player connections.
func (s *Server) Listen(port int) error {
	if s.handleConnection == nil {
		return errors.New("A connection handler must be specified before Listen() is called.")
	}
	l, err := net.Listen("tcp", ":4000")
	if err != nil {
		return err
	}
	s.ln = l

	// Create a channel for incoming connections.
	newUserChan := make(chan net.Conn, 1)
	go func() {
		for {
			c, err := l.Accept()
			if err != nil {
				log.Error().Err(err)
				continue
			}
			newUserChan <- c
		}
	}()

	// Loop and select
	for {
		select {
		case c := <-newUserChan:
			go s.handleConnection(c)
		case <-util.SigIntChannel():
			l.Close()
			return nil
		}
	}
}
