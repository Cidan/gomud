package main

import (
	"net/http"
	_ "net/http/pprof"

	"github.com/Cidan/gomud/construct"
	"github.com/Cidan/gomud/server"
	"github.com/rs/zerolog/log"
)

func main() {

	err := construct.LoadRooms()
	if err != nil {
		panic(err)
	}

	if construct.WorldSize() == 0 {
		makeDefaultRoomSet()
	}

	server := server.New()
	go startDebugServer()

	log.Info().Msg("Gomud listening on port 4000.")
	if err := server.Listen(4000); err != nil {
		log.Panic().Err(err).Msg("Error while listening for new connections.")
	}
	log.Info().Msg("Server shutting down.")
}

func startDebugServer() {
	err := http.ListenAndServe(":8472", nil)
	if err != nil {
		panic(err)
	}
}

func makeDefaultRoomSet() {
	room := construct.NewRoom()
	room.SetName("The Alpha")
	room.SetDescription("It all starts here.")
	room.SetCoordinates(0, 0, 0)

	err := room.Save()
	if err != nil {
		panic(err)
	}
	construct.AddRoom(room)

	room = construct.NewRoom()
	room.SetName("The Omega")
	room.SetDescription("It all ends here.")
	room.SetCoordinates(1, 0, 0)

	err = room.Save()
	if err != nil {
		panic(err)
	}
	construct.AddRoom(room)
}
