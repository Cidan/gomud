package main

import (
	"github.com/Cidan/gomud/atlas"
	"github.com/Cidan/gomud/room"
	"github.com/Cidan/gomud/server"
	"github.com/rs/zerolog/log"
)

func main() {
	err := atlas.LoadWorld()
	if err != nil {
		panic(err)
	}

	makeDefaultRoom()
	server := server.New()
	log.Info().Msg("Gomud listening on port 4000.")
	if err := server.Listen(4000); err != nil {
		log.Panic().Err(err).Msg("Error while listening for new connections.")
	}
	log.Info().Msg("Server shutting down.")
}

func makeDefaultRoom() {
	atlas.AddRoom(room.New(&room.Data{
		Name:        "The Alpha",
		Description: "It all starts here.",
		X:           0,
		Y:           0,
		Z:           0,
	}))
	atlas.AddRoom(room.New(&room.Data{
		Name:        "The Omega",
		Description: "It all ends here.",
		X:           1,
		Y:           0,
		Z:           0,
	}))
}
