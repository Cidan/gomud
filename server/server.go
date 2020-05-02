package server

import (
	"net"

	"github.com/Cidan/gomud/player"
	"github.com/Cidan/gomud/util"

	"github.com/rs/zerolog/log"
)

// Server is a network server construct
// that handles incoming player connections.
type Server struct {
	listener net.Listener
}

// New Server
func New() *Server {
	return &Server{}
}

func (s *Server) handleConnection(c net.Conn) {
	log.Info().
		Str("address", c.RemoteAddr().String()).
		Msg("New connection")
	p := player.New()
	// This blocks as it starts the interp loop
	p.SetConnection(c)
}

// Listen on a port for player connections.
func (s *Server) Listen(port int) error {
	l, err := net.Listen("tcp", ":8090")
	if err != nil {
		return err
	}
	s.listener = l

	// Handle our exit
	go s.onSigInt()

	// Loop for new connections
	for {
		c, err := l.Accept()
		if err != nil {
			break
		} else {
			go s.handleConnection(c)
			continue
		}
	}
	return nil
}

func (s *Server) onSigInt() {
	<-util.SigIntChannel()
	s.listener.Close()
}
