package main

import (
	"context"
	"net/http"
	_ "net/http/pprof"
	"os"
	"runtime"
	"time"

	"github.com/Cidan/gomud/construct"
	"github.com/Cidan/gomud/lock"
	"github.com/Cidan/gomud/server"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	if os.Getenv("MUDDEBUG") != "" {
		go startDebugServer()
		go startGC()
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}
	ctx := lock.Context(context.Background(), "main")
	err := construct.LoadRooms(ctx)
	if err != nil {
		panic(err)
	}

	if construct.Atlas.WorldSize() == 0 {
		construct.Atlas.MakeDefaultRoomSet(ctx)
	}

	server := server.New()

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

func startGC() {
	for {
		runtime.GC()
		time.Sleep(5 * time.Second)
	}
}
