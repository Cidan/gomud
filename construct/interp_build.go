package construct

import (
	"strings"

	"github.com/rs/zerolog/log"
)

// BuildInterp is the builder interp, used for world crafting and modifying
// the permanent game world.
type BuildInterp struct {
	p        *Player
	commands *commandMap
}

// NewBuildInterp creates a new build interp.
func NewBuildInterp(p *Player) *BuildInterp {
	b := &BuildInterp{
		p: p,
	}

	commands := newCommands()
	commands.Add(&command{
		name: "dig",
		Fn:   b.DoDig,
	})
	b.commands = commands
	return b
}

func (b *BuildInterp) Read(text string) error {
	all := strings.SplitN(text, " ", 2)
	log.
		Debug().
		Interface("command", all).
		Str("player.uuid", b.p.GetUUID()).
		Str("player.name", b.p.GetName()).
		Msg("Command")

	if b.commands.Has(all[0]) {
		return b.commands.Process(all[0], all[1:]...)
	}
	return b.p.gameInterp.commands.Process(all[0], all[1:]...)
}

func (b *BuildInterp) doDigDir(dir string) error {
	b.p.Write("Not yet implemented.\n")
	return nil
}

// DoDig will create a new room in the direction the player specifies.
func (b *BuildInterp) DoDig(args ...string) error {
	if len(args) == 0 || args[0] == "" {
		b.p.Write("Which direction do you want to dig?")
		return nil
	}
	switch args[0] {
	case "north", "n":
		return b.doDigDir("north")
	case "east", "e":
		return b.doDigDir("east")
	case "south", "s":
		return b.doDigDir("south")
	case "west", "w":
		return b.doDigDir("west")
	case "up", "u":
		return b.doDigDir("up")
	case "down", "d":
		return b.doDigDir("down")
	default:
		b.p.Write("That's not a valid direction to dig in.")
		return nil
	}
}
