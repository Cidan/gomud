package main

import (
	"github.com/Cidan/gomud/atlas"
	"github.com/Cidan/gomud/construct"
	"github.com/Cidan/gomud/server"
	"github.com/rs/zerolog/log"
)

func main() {
	err := atlas.SetupWorld()
	if err != nil {
		panic(err)
	}

	err = construct.LoadRooms()
	if err != nil {
		panic(err)
	}

	if atlas.WorldSize() == 0 {
		makeDefaultRoomSet()
	}

	server := server.New()
	log.Info().Msg("Gomud listening on port 4000.")
	if err := server.Listen(8090); err != nil {
		log.Panic().Err(err).Msg("Error while listening for new connections.")
	}
	log.Info().Msg("Server shutting down.")
}

func makeDefaultRoomSet() {
	room := construct.NewRoom(&construct.RoomData{
		Name:        "The Alpha",
		Description: "It all starts here.",
		X:           0,
		Y:           0,
		Z:           0,
	})
	err := room.Save()
	if err != nil {
		panic(err)
	}
	atlas.AddRoom(room)
	room = construct.NewRoom(&construct.RoomData{
		Name:        "The Omega",
		Description: "It all ends here.",
		X:           1,
		Y:           0,
		Z:           0,
	})
	err = room.Save()
	if err != nil {
		panic(err)
	}
	atlas.AddRoom(room)
}
