package construct

import (
	"context"
	"strings"
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
	}).Add(&command{
		name: "build",
		Fn:   b.DoBuild,
	}).Add(&command{
		name: "autobuild",
		Fn:   b.Autobuild,
	}).Add(&command{
		name: "set",
		Fn:   b.DoSet,
	}).Add(&command{
		name:  "north",
		alias: []string{"n"},
		Fn:    b.DoNorth,
	}).Add(&command{
		name:  "east",
		alias: []string{"e"},
		Fn:    b.DoEast,
	}).Add(&command{
		name:  "south",
		alias: []string{"s"},
		Fn:    b.DoSouth,
	}).Add(&command{
		name:  "west",
		alias: []string{"w"},
		Fn:    b.DoWest,
	}).Add(&command{
		name:  "up",
		alias: []string{"u"},
		Fn:    b.DoUp,
	}).Add(&command{
		name:  "down",
		alias: []string{"d"},
		Fn:    b.DoDown,
	}).Add(&command{
		name: "edit",
		Fn:   b.DoEdit,
	})
	b.commands = commands
	return b
}

func (b *BuildInterp) Read(ctx context.Context, text string) error {
	all := strings.SplitN(text, " ", 2)
	/*
		log.
			Debug().
			Interface("command", all).
			Str("player.uuid", b.p.GetUUID()).
			Str("player.name", b.p.GetName()).
			Msg("Command")
	*/
	if b.commands.Has(all[0]) {
		return b.commands.Process(ctx, all[0], all[1:]...)
	}
	return b.p.gameInterp.commands.Process(ctx, all[0], all[1:]...)
}

func (b *BuildInterp) doDigDir(dir direction) error {
	currentRoom := b.p.GetRoom()
	if currentRoom.PhysicalRoom(dir) != nil {
		b.p.Write("There's already a room '%s'.\n", Atlas.dirToName(dir))
		return nil
	}

	rX, rY, rZ := Atlas.getRelativeDir(dir)

	room := NewRoom()
	room.Data.X = currentRoom.Data.X + rX
	room.Data.Y = currentRoom.Data.Y + rY
	room.Data.Z = currentRoom.Data.Z + rZ

	for _, exitDir := range exitDirections {
		if inverseDirections[dir] == exitDir {
			room.SetExitRoom(exitDir, currentRoom)
			continue
		}
		room.Exit(exitDir).Wall = true
	}
	currentRoom.Exit(dir).Wall = false
	currentRoom.SetExitRoom(dir, room)

	if err := room.Save(); err != nil {
		return err
	}

	if err := currentRoom.Save(); err != nil {
		return err
	}

	Atlas.AddRoom(room)

	b.p.gameInterp.doDir(dir)
	return nil
}

// DoDig will create a new room in the direction the player specifies.
func (b *BuildInterp) DoDig(ctx context.Context, args ...string) error {
	if len(args) == 0 || args[0] == "" {
		b.p.Write("Which direction do you want to dig?")
		return nil
	}
	switch args[0] {
	case "north", "n":
		return b.doDigDir(dirNorth)
	case "east", "e":
		return b.doDigDir(dirEast)
	case "south", "s":
		return b.doDigDir(dirSouth)
	case "west", "w":
		return b.doDigDir(dirWest)
	case "up", "u":
		return b.doDigDir(dirUp)
	case "down", "d":
		return b.doDigDir(dirDown)
	default:
		b.p.Write("That's not a valid direction to dig in.")
		return nil
	}
}

// DoBuild deactivates build mode.
func (b *BuildInterp) DoBuild(ctx context.Context, args ...string) error {
	b.p.Game(ctx)
	b.p.Write("Build mode deactivated.")
	return nil
}

// Autobuild enables autobuild, which will automatically cause the player
// to dig in the direction of their movement.
func (b *BuildInterp) Autobuild(ctx context.Context, args ...string) error {
	if v := b.p.ToggleFlag("autobuild"); v {
		b.p.Write("Autobuild has been enabled.")
	} else {
		b.p.Write("Autobuild has been disabled.")
	}
	return nil
}

func (b *BuildInterp) DoNorth(ctx context.Context, args ...string) error {
	p := b.p
	g := p.gameInterp
	if !p.Flag("autobuild") || p.GetRoom().PhysicalRoom(dirNorth) != nil {
		g.doDir(dirNorth)
		return nil
	}
	return b.doDigDir(dirNorth)
}
func (b *BuildInterp) DoSouth(ctx context.Context, args ...string) error {
	p := b.p
	g := p.gameInterp
	if !p.Flag("autobuild") || p.GetRoom().PhysicalRoom(dirSouth) != nil {
		g.doDir(dirSouth)
		return nil
	}
	return b.doDigDir(dirSouth)
}
func (b *BuildInterp) DoEast(ctx context.Context, args ...string) error {
	p := b.p
	g := p.gameInterp
	if !p.Flag("autobuild") || p.GetRoom().PhysicalRoom(dirEast) != nil {
		g.doDir(dirEast)
		return nil
	}
	return b.doDigDir(dirEast)
}
func (b *BuildInterp) DoWest(ctx context.Context, args ...string) error {
	p := b.p
	g := p.gameInterp
	if !p.Flag("autobuild") || p.GetRoom().PhysicalRoom(dirWest) != nil {
		g.doDir(dirWest)
		return nil
	}
	return b.doDigDir(dirWest)
}
func (b *BuildInterp) DoUp(ctx context.Context, args ...string) error {
	p := b.p
	g := p.gameInterp
	if !p.Flag("autobuild") || p.GetRoom().PhysicalRoom(dirUp) != nil {
		g.doDir(dirUp)
		return nil
	}
	return b.doDigDir(dirUp)
}
func (b *BuildInterp) DoDown(ctx context.Context, args ...string) error {
	p := b.p
	g := p.gameInterp
	if !p.Flag("autobuild") || p.GetRoom().PhysicalRoom(dirDown) != nil {
		g.doDir(dirDown)
		return nil
	}
	return b.doDigDir(dirDown)
}

func (b *BuildInterp) DoSet(ctx context.Context, args ...string) error {
	if len(args) == 0 {
		b.p.Write("Set what?")
		return nil
	}
	args = strings.SplitN(args[0], " ", 2)
	switch args[0] {
	case "room":
		return b.setRoom(args[1:]...)
	default:
		b.p.Write("No such thing to set.")
		return nil
	}
}

func (b *BuildInterp) setRoom(args ...string) error {
	p := b.p
	room := p.GetRoom()

	if len(args) == 0 {
		p.Write("What do you want to set on the room?")
		return nil
	}
	args = strings.SplitN(args[0], " ", 2)
	switch args[0] {
	case "name":
		if len(args) < 2 {
			p.Write("What do you want to set the description to?")
			return nil
		}
		room.SetName(args[1])
		p.Write("Name set.")
		return room.Save()
	case "description":
		if len(args) < 2 {
			p.Write("What do you want to set the description to?")
			return nil
		}
		room.SetDescription(args[1])
		p.Write("Description set.")
		return room.Save()
	default:
		p.Write("There's no such room property to set.")
		return nil
	}

}

func (b *BuildInterp) DoEdit(ctx context.Context, args ...string) error {
	if len(args) == 0 {
		b.p.Write("What would you like to edit?")
		return nil
	}
	sp := strings.SplitN(args[0], " ", 3)
	if len(sp) < 2 {
		b.p.Write("What would you like to edit?")
		return nil
	}
	switch sp[0] {
	case "room":
		return b.editRoom(ctx, sp[1])
	}
	return nil
}

func (b *BuildInterp) editRoom(ctx context.Context, field string) error {
	p := b.p
	room := p.GetRoom()

	switch field {
	case "name":
		ctx = p.textInterp.Start(&room.Data.Name)
	case "description":
		ctx = p.textInterp.Start(&room.Data.Description)
	}

	p.setInterp(ctx, p.textInterp)
	p.Write("You are now editing text. Type :q to quit, :w to save, and :? for help.")
	// TODO(lobato): figure out how to do this.
	/*
		go func(ctx context.Context, room *Room) {
			<-ctx.Done()
			room.Save()
			p.setInterp(ctx, p.buildInterp)
			p.Command("look")
		}(ctx, room)
	*/
	return nil
}
