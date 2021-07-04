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
	currentRoom := b.p.GetRoom()
	if currentRoom.LinkedRoom(dir) != nil {
		b.p.Write("There's already a room '%s'.\n", dir)
		return nil
	}
	rX, rY, rZ := b.getRelativeDir(dir)
	room := NewRoom(&RoomData{
		Name:        "New Room",
		Description: "This is a new room, with a new description.",
		X:           currentRoom.Data.X + rX,
		Y:           currentRoom.Data.Y + rY,
		Z:           currentRoom.Data.Z + rZ,
	})

	if err := room.Save(); err != nil {
		return err
	}
	AddRoom(room)

	b.p.ToRoom(room)
	b.p.Command("look")
	return nil
}

func (b *BuildInterp) getRelativeDir(dir string) (x, y, z int64) {
	switch dir {
	case "north":
		return 0, 1, 0
	case "south":
		return 0, -1, 0
	case "east":
		return 1, 0, 0
	case "west":
		return -1, 0, 0
	case "up":
		return 0, 0, 1
	case "down":
		return 0, 0, -1
	}
	return 0, 0, 0
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
