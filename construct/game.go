package construct

import (
	"strings"

	"github.com/Cidan/gomud/types"
	"github.com/rs/zerolog/log"
)

// Game interp for handling user login
type Game struct {
	p types.Player
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
	}).Add(&command{
		name: "build",
		Fn:   DoBuild,
	}).Add(&command{
		name:  "north",
		alias: []string{"n"},
		Fn:    DoNorth,
	}).Add(&command{
		name:  "east",
		alias: []string{"e"},
		Fn:    DoEast,
	}).Add(&command{
		name:  "south",
		alias: []string{"s"},
		Fn:    DoSouth,
	}).Add(&command{
		name:  "west",
		alias: []string{"w"},
		Fn:    DoWest,
	}).Add(&command{
		name:  "up",
		alias: []string{"u"},
		Fn:    DoUp,
	}).Add(&command{
		name:  "down",
		alias: []string{"d"},
		Fn:    DoDown,
	})
}

// NewGame interp for a player. This is the main game state interp
// for which all gameplay commands are run.
func NewGame(p types.Player) *Game {
	g := &Game{
		p: p,
	}
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
	return gameCommands.Process(g.p, all[0], all[1:]...)
}

// Commands go under here.

// DoLook Look at the current room, an object, a player, or an NPC
func DoLook(p types.Player, args ...string) error {
	room := p.GetRoom()
	p.Buffer("\n\n%s\n\n", room.GetName())
	p.Buffer("  %s\n", room.GetDescription())
	p.Flush()
	return nil
}

// DoSave will save a player to durable storage.
func DoSave(p types.Player, args ...string) error {
	err := p.Save()
	if err == nil {
		p.Write("Your player has been saved.")
	}
	return err
}

// DoQuit will exit the player from the game world.
func DoQuit(p types.Player, args ...string) error {
	p.Write("See ya!\n")
	p.Stop()
	return nil
}

// DoBuild enables build mode for the player.
func DoBuild(p types.Player, args ...string) error {
	p.Write("Entering build mode\n")
	p.SetInterp(NewBuild(p))
	return nil
}

// doDir for moving a player in a direction or through a portal.
func doDir(p types.Player, dir string) {
	target := p.GetRoom().LinkedRoom(dir)
	if target != nil {
		p.ToRoom(target)
		p.Command("look")
		return
	}
	p.Write("You can't go that way!")
	return
}

func DoNorth(p types.Player, args ...string) error {
	doDir(p, "north")
	return nil
}

func DoEast(p types.Player, args ...string) error {
	doDir(p, "east")
	return nil
}

func DoSouth(p types.Player, args ...string) error {
	doDir(p, "south")
	return nil
}

func DoWest(p types.Player, args ...string) error {
	doDir(p, "west")
	return nil
}

func DoUp(p types.Player, args ...string) error {
	doDir(p, "up")
	return nil
}
func DoDown(p types.Player, args ...string) error {
	doDir(p, "down")
	return nil
}
