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
	construct.AddRoom(room)
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
	construct.AddRoom(room)
}
