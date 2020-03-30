package interp

import (
	"strings"

	"github.com/Cidan/gomud/player"

	"github.com/rs/zerolog/log"
)

// Game interp for handling user login
type Game struct {
	p *player.Player
}

var GameCommands *CommandMap

func init() {
	GameCommands = NewCommands()
	GameCommands.Add(&Command{
		name:  "look",
		alias: []string{"l"},
		Fn:    DoLook,
	}).Add(&Command{
		name: "save",
		Fn:   DoSave,
	}).Add(&Command{
		name: "quit",
		Fn:   DoQuit,
	})
}

func NewGame(p *player.Player) *Game {
	g := &Game{p: p}
	return g
}

func (g *Game) Read(text string) error {
	all := strings.SplitN(text, " ", 2)
	log.Debug().Interface("command", all).Str("player", g.p.Data.UUID).Msg("Command")
	return GameCommands.Process(g.p, all[0], all[1:]...)
}

// Commands go under here.

// DoLook Look at the current room, an object, a player, or an NPC
func DoLook(p *player.Player, args ...string) error {
	p.Write("You can't see anything. %s", args)
	return nil
}

func DoSave(p *player.Player, args ...string) error {
	err := p.Save()
	if err == nil {
		p.Write("Your player has been saved.")
	}
	return err
}

func DoQuit(p *player.Player, args ...string) error {
	p.Write("See ya!\n")
	p.Stop()
	return nil
}
