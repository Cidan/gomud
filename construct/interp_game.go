package construct

import (
	"context"
	"fmt"
	"strconv"
	"strings"
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
	}).Add(&command{
		name: "prompt",
		Fn:   g.DoPrompt,
	}).Add(&command{
		name: "color",
		Fn:   g.DoColor,
	}).Add(&command{
		name: "say",
		Fn:   g.DoSay,
	}).Add(&command{
		name: "map",
		Fn:   g.DoMap,
	})

	g.commands = commands
	return g
}

func (g *Game) Read(ctx context.Context, text string) error {
	all := strings.SplitN(text, " ", 2)
	/*
		log.
			Debug().
			Interface("command", all).
			Str("player.uuid", g.p.GetUUID()).
			Str("player.name", g.p.GetName()).
			Msg("Command")
	*/
	return g.commands.Process(ctx, all[0], all[1:]...)
}

// Commands go under here.

// DoLook Look at the current room, an object, a player, or an NPC
func (g *Game) DoLook(ctx context.Context, args ...string) error {
	room := g.p.GetRoom(ctx)

	// Display the room name.
	g.p.Buffer(ctx, "\n\n%s\n", room.GetName())

	// Display exits.
	var exitFound bool
	g.p.Buffer(ctx, "{c[Exits:")
	for _, dir := range exitDirections {
		if g.p.CanExit(ctx, dir) {
			g.p.Buffer(ctx, " %s", Atlas.dirToName(dir))
			exitFound = true
		}
	}
	if !exitFound {
		g.p.Buffer(ctx, " none")
	}
	g.p.Buffer(ctx, "]{x\n")

	// Display the automap if the player has it enabled.
	if g.p.Flag(ctx, "automap") {
		g.p.Buffer(ctx, "\n%s\n\n", g.p.Map(ctx, 5))
	} else {
		g.p.Buffer(ctx, "\n")
	}

	// Show the room description.
	g.p.Buffer(ctx, "  %s\n", room.GetDescription())

	// List all the players in the room.
	room.AllPlayers(func(uuid string, rp *Player) {
		if rp == g.p {
			return
		}
		g.p.Buffer(ctx, "\n%s\n", rp.PlayerDescription())
	})

	// Flush our buffered output to the player.
	g.p.Flush(ctx)
	return nil
}

// DoSave will save a player to durable storage.
func (g *Game) DoSave(ctx context.Context, args ...string) error {
	err := g.p.Save()
	if err == nil {
		g.p.Write(ctx, "Your player has been saved.")
	}
	return err
}

// DoQuit will exit the player from the game world.
func (g *Game) DoQuit(ctx context.Context, args ...string) error {
	g.p.Write(ctx, "See ya!\n")
	g.p.Stop(ctx)
	return nil
}

// DoBuild enables build mode for the player.
func (g *Game) DoBuild(ctx context.Context, args ...string) error {
	g.p.Build(ctx)
	g.p.Write(ctx, "Entering build mode.")
	return nil
}

// doDir for moving a player in a direction or through a portal.
func (g *Game) doDir(ctx context.Context, dir direction) {
	room := g.p.GetRoom(ctx)

	if g.p.CanExit(ctx, dir) {
		target := room.LinkedRoom(dir)
		g.p.ToRoom(ctx, target)
		g.p.Command("look")
		return
	}

	if room.IsExitClosed(dir) {
		g.p.Write(ctx, "The exit %s is closed!", Atlas.dirToName(dir))
		return
	}

	g.p.Write(ctx, "You can't go that way!")
	return
}

// DoNorth moves the player north.
func (g *Game) DoNorth(ctx context.Context, args ...string) error {
	g.doDir(ctx, dirNorth)
	return nil
}

// DoEast moves the player east.
func (g *Game) DoEast(ctx context.Context, args ...string) error {
	g.doDir(ctx, dirEast)
	return nil
}

// DoSouth moves the player south.
func (g *Game) DoSouth(ctx context.Context, args ...string) error {
	g.doDir(ctx, dirSouth)
	return nil
}

// DoWest moves the player west.
func (g *Game) DoWest(ctx context.Context, args ...string) error {
	g.doDir(ctx, dirWest)
	return nil
}

// DoUp moves the player up.
func (g *Game) DoUp(ctx context.Context, args ...string) error {
	g.doDir(ctx, dirUp)
	return nil
}

// DoDown moves the player down.
func (g *Game) DoDown(ctx context.Context, args ...string) error {
	g.doDir(ctx, dirDown)
	return nil
}

// DoSay will send a message to all players in the local room.
func (g *Game) DoSay(ctx context.Context, args ...string) error {
	p := g.p

	if len(args) == 0 {
		p.Write(ctx, "Say what?")
		return nil
	}

	room := p.GetRoom(ctx)
	if room == nil {
		return fmt.Errorf("player %s not in a valid room", p.GetName())
	}
	text := strings.Join(args, " ")
	room.AllPlayers(func(uuid string, rp *Player) {
		if rp == p {
			rp.Write(ctx, "{yYou say, {x'%s{x'", text)
			return
		}
		rp.Write(ctx, "{y%s says, {x'%s{x'", p.GetName(), text)
	})
	return nil
}

// DoPrompt will either enable/disable a user prompt, or set the prompt string.
func (g *Game) DoPrompt(ctx context.Context, args ...string) error {
	if len(args) == 0 {
		if v := g.p.ToggleFlag(ctx, "prompt"); v {
			g.p.Write(ctx, "Prompt enabled.")
		} else {
			g.p.Write(ctx, "Prompt disabled.")
		}
		return nil
	}
	g.p.SetPrompt(strings.Join(args, " "))
	g.p.Write(ctx, "Prompt set.")
	return nil
}

// DoColor will toggle the color flag for a player.
func (g *Game) DoColor(ctx context.Context, args ...string) error {
	if v := g.p.ToggleFlag(ctx, "color"); v {
		g.p.Write(ctx, "{gColor enabled!{x")
	} else {
		g.p.Write(ctx, "Color disabled :(")
	}
	return nil
}

// DoMap will display a map with a given radius around the player.
func (g *Game) DoMap(ctx context.Context, args ...string) error {
	var radius int64
	p := g.p
	if len(args) > 0 {
		switch {
		case args[0] == "off":
			p.DisableFlag(ctx, "automap")
			p.Write(ctx, "Automap turned off.")
			return nil
		case args[0] == "on":
			p.EnableFlag(ctx, "automap")
			p.Write(ctx, "Automap turned on.")
			return nil
		default:
			r, err := strconv.Atoi(args[0])
			if err != nil {
				p.Write(ctx, "You must specify a number for your map size, i.e. 'map 5'")
			}
			radius = int64(r)
		}
	}
	if radius > 100 || radius == 0 {
		radius = int64(100)
	}
	g.p.Write(ctx, p.Map(ctx, radius))
	return nil
}
