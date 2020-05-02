package interp

import (
	"strings"

	"github.com/rs/zerolog/log"
)

// Game interp for handling user login
type Game struct {
	p player
}

var gameCommands *commandMap

func init() {
	gameCommands = newCommands()
	gameCommands.Add(&command{
		name:  "look",
		alias: []string{"l"},
		Fn:    DoLook,
	}).Add(&command{
		name: "save",
		Fn:   DoSave,
	}).Add(&command{
		name: "quit",
		Fn:   DoQuit,
	})
}

// NewGame interp for a player. This is the main game state interp
// for which all gameplay commands are run.
func NewGame(p player) *Game {
	g := &Game{
		p: p,
	}
	return g
}

func (g *Game) Read(text string) error {
	all := strings.SplitN(text, " ", 2)
	log.Debug().Interface("command", all).Str("player", g.p.GetUUID()).Msg("Command")
	return gameCommands.Process(g.p, all[0], all[1:]...)
}

// Commands go under here.

// DoLook Look at the current room, an object, a player, or an NPC
func DoLook(p player, args ...string) error {
	p.Write("You can't see anything. %s", args)
	return nil
}

// DoSave will save a player to durable storage.
func DoSave(p player, args ...string) error {
	err := p.Save()
	if err == nil {
		p.Write("Your player has been saved.")
	}
	return err
}

// DoQuit will exit the player from the game world.
func DoQuit(p player, args ...string) error {
	p.Write("See ya!\n")
	p.Stop()
	return nil
}
