package main

import (
	"net"

	"github.com/Cidan/gomud/atlas"
	"github.com/Cidan/gomud/player"
	"github.com/Cidan/gomud/server"
	"github.com/Cidan/gomud/world"
	"github.com/rs/zerolog/log"
)

func main() {
	atlas.SetupWorld()
	atlas.SetupPlayer()
	MakeDefaultRoom()
	server := server.New()

	server.SetHandler(func(c net.Conn) {
		log.Info().
			Str("address", c.RemoteAddr().String()).
			Msg("New connection")
		p := player.New()
		p.SetConnection(c)
		p.Write("Welcome, by what name are you known?")
		atlas.StartPlayer(p)
	})

	log.Info().Msg("Gomud listening on port 4000.")
	if err := server.Listen(4000); err != nil {
		log.Panic().Err(err).Msg("Error while listening for new connections.")
	}
	log.Info().Msg("Server shutting down.")
}

func MakeDefaultRoom() {
	r := world.NewRoom(&world.RoomData{
		Name:        "The Alpha",
		Description: "It all starts here.",
		X:           0,
		Y:           0,
		Z:           0,
	})
	atlas.AddRoom(r)
}
