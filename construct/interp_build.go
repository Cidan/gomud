package construct

import (
	"strings"

	"github.com/rs/zerolog/log"
)

// BuildInterp is the builder interp, used for world crafting and modifying
// the permanent game world.
type BuildInterp struct {
	p *Player
}

var buildCommands *commandMap

func init() {
	buildCommands = newCommands()
	buildCommands.Add(&command{
		name: "dig",
		Fn:   DoDig,
	})
}

// NewBuildInterp creates a new build interp.
func NewBuildInterp(p *Player) *BuildInterp {
	b := &BuildInterp{
		p: p,
	}
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

	if buildCommands.Has(all[0]) {
		return buildCommands.Process(b.p, all[0], all[1:]...)
	}
	return b.p.gameInterp.commands.Process(b.p, all[0], all[1:]...)
}

func doDigDir(p *Player, dir string) error {
	p.Write("Not yet implemented.\n")
	return nil
}

// DoDig will create a new room in the direction the player specifies.
func DoDig(p *Player, args ...string) error {
	if len(args) == 0 || args[0] == "" {
		p.Write("Which direction do you want to dig?")
		return nil
	}
	switch args[0] {
	case "north", "n":
		return doDigDir(p, "north")
	case "east", "e":
		return doDigDir(p, "east")
	case "south", "s":
		return doDigDir(p, "south")
	case "west", "w":
		return doDigDir(p, "west")
	case "up", "u":
		return doDigDir(p, "up")
	case "down", "d":
		return doDigDir(p, "down")
	default:
		p.Write("That's not a valid direction to dig in.")
		return nil
	}
}
