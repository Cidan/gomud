package construct

import (
	"strings"

	"github.com/rs/zerolog/log"
)

// Game interp for handling user login
type Game struct {
	p        *Player
	commands *commandMap
}

// NewGameInterp interp for a player. This is the main game state interp
// for which all gameplay commands are run.
func NewGameInterp(p *Player) *Game {
	g := &Game{
		p: p,
	}

	commands := newCommands()
	commands.Add(&command{
		name:  "look",
		alias: []string{"l"},
		Fn:    g.DoLook,
	}).Add(&command{
		name: "save",
		Fn:   g.DoSave,
	}).Add(&command{
		name: "quit",
		Fn:   g.DoQuit,
	}).Add(&command{
		name: "build",
		Fn:   g.DoBuild,
	}).Add(&command{
		name:  "north",
		alias: []string{"n"},
		Fn:    g.DoNorth,
	}).Add(&command{
		name:  "east",
		alias: []string{"e"},
		Fn:    g.DoEast,
	}).Add(&command{
		name:  "south",
		alias: []string{"s"},
		Fn:    g.DoSouth,
	}).Add(&command{
		name:  "west",
		alias: []string{"w"},
		Fn:    g.DoWest,
	}).Add(&command{
		name:  "up",
		alias: []string{"u"},
		Fn:    g.DoUp,
	}).Add(&command{
		name:  "down",
		alias: []string{"d"},
		Fn:    g.DoDown,
	})

	g.commands = commands
	return g
}

func (g *Game) Read(text string) error {
	all := strings.SplitN(text, " ", 2)
	log.
		Debug().
		Interface("command", all).
		Str("player.uuid", g.p.GetUUID()).
		Str("player.name", g.p.GetName()).
		Msg("Command")
	return g.commands.Process(all[0], all[1:]...)
}

// Commands go under here.

// DoLook Look at the current room, an object, a player, or an NPC
func (g *Game) DoLook(args ...string) error {
	room := g.p.GetRoom()
	g.p.Buffer("\n\n%s\n\n", room.GetName())
	g.p.Buffer("  %s\n", room.GetDescription())
	g.p.Flush()
	return nil
}

// DoSave will save a player to durable storage.
func (g *Game) DoSave(args ...string) error {
	err := g.p.Save()
	if err == nil {
		g.p.Write("Your player has been saved.")
	}
	return err
}

// DoQuit will exit the player from the game world.
func (g *Game) DoQuit(args ...string) error {
	g.p.Write("See ya!\n")
	g.p.Stop()
	return nil
}

// DoBuild enables build mode for the player.
func (g *Game) DoBuild(args ...string) error {
	g.p.Write("Entering build mode\n")
	g.p.Build()
	return nil
}

// doDir for moving a player in a direction or through a portal.
func (g *Game) doDir(dir string) {
	target := g.p.GetRoom().LinkedRoom(dir)
	if target != nil {
		g.p.ToRoom(target)
		g.p.Command("look")
		return
	}
	g.p.Write("You can't go that way!")
	return
}

// DoNorth moves the player north.
func (g *Game) DoNorth(args ...string) error {
	g.doDir("north")
	return nil
}

// DoEast moves the player east.
func (g *Game) DoEast(args ...string) error {
	g.doDir("east")
	return nil
}

// DoSouth moves the player south.
func (g *Game) DoSouth(args ...string) error {
	g.doDir("south")
	return nil
}

// DoWest moves the player west.
func (g *Game) DoWest(args ...string) error {
	g.doDir("west")
	return nil
}

// DoUp moves the player up.
func (g *Game) DoUp(args ...string) error {
	g.doDir("up")
	return nil
}

// DoDown moves the player down.
func (g *Game) DoDown(args ...string) error {
	g.doDir("down")
	return nil
}
