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
	room := g.p.GetRoom()

	// Display the room name.
	g.p.Buffer("\n\n%s\n", room.GetName())

	// Display exits.
	var exitFound bool
	g.p.Buffer("{c[Exits:")
	for _, dir := range exitDirections {
		if g.p.CanExit(dir) {
			g.p.Buffer(" %s", Atlas.dirToName(dir))
			exitFound = true
		}
	}
	if !exitFound {
		g.p.Buffer(" none")
	}
	g.p.Buffer("]{x\n")

	// Display the automap if the player has it enabled.
	if g.p.Flag("automap") {
		g.p.Buffer("\n%s\n\n", g.p.Map(5))
	} else {
		g.p.Buffer("\n")
	}

	// Show the room description.
	g.p.Buffer("  %s\n", room.GetDescription())

	// List all the players in the room.
	room.AllPlayers(func(uuid string, rp *Player) {
		if rp == g.p {
			return
		}
		g.p.Buffer("\n%s\n", rp.PlayerDescription())
	})

	// Flush our buffered output to the player.
	g.p.Flush()
	return nil
}

// DoSave will save a player to durable storage.
func (g *Game) DoSave(ctx context.Context, args ...string) error {
	err := g.p.Save()
	if err == nil {
		g.p.Write("Your player has been saved.")
	}
	return err
}

// DoQuit will exit the player from the game world.
func (g *Game) DoQuit(ctx context.Context, args ...string) error {
	g.p.Write("See ya!\n")
	g.p.Stop()
	return nil
}

// DoBuild enables build mode for the player.
func (g *Game) DoBuild(ctx context.Context, args ...string) error {
	g.p.Build(ctx)
	g.p.Write("Entering build mode.")
	return nil
}

// doDir for moving a player in a direction or through a portal.
func (g *Game) doDir(dir direction) {
	room := g.p.GetRoom()

	if g.p.CanExit(dir) {
		target := room.LinkedRoom(dir)
		g.p.ToRoom(target)
		g.p.Command("look")
		return
	}

	if room.IsExitClosed(dir) {
		g.p.Write("The exit %s is closed!", Atlas.dirToName(dir))
		return
	}

	g.p.Write("You can't go that way!")
	return
}

// DoNorth moves the player north.
func (g *Game) DoNorth(ctx context.Context, args ...string) error {
	g.doDir(dirNorth)
	return nil
}

// DoEast moves the player east.
func (g *Game) DoEast(ctx context.Context, args ...string) error {
	g.doDir(dirEast)
	return nil
}

// DoSouth moves the player south.
func (g *Game) DoSouth(ctx context.Context, args ...string) error {
	g.doDir(dirSouth)
	return nil
}

// DoWest moves the player west.
func (g *Game) DoWest(ctx context.Context, args ...string) error {
	g.doDir(dirWest)
	return nil
}

// DoUp moves the player up.
func (g *Game) DoUp(ctx context.Context, args ...string) error {
	g.doDir(dirUp)
	return nil
}

// DoDown moves the player down.
func (g *Game) DoDown(ctx context.Context, args ...string) error {
	g.doDir(dirDown)
	return nil
}

// DoSay will send a message to all players in the local room.
func (g *Game) DoSay(ctx context.Context, args ...string) error {
	p := g.p

	if len(args) == 0 {
		p.Write("Say what?")
		return nil
	}

	room := p.GetRoom()
	if room == nil {
		return fmt.Errorf("player %s not in a valid room", p.GetName())
	}
	text := strings.Join(args, " ")
	room.AllPlayers(func(uuid string, rp *Player) {
		if rp == p {
			rp.Write("{yYou say, {x'%s{x'", text)
			return
		}
		rp.Write("{y%s says, {x'%s{x'", p.GetName(), text)
	})
	return nil
}

// DoPrompt will either enable/disable a user prompt, or set the prompt string.
func (g *Game) DoPrompt(ctx context.Context, args ...string) error {
	if len(args) == 0 {
		if v := g.p.ToggleFlag("prompt"); v {
			g.p.Write("Prompt enabled.")
		} else {
			g.p.Write("Prompt disabled.")
		}
		return nil
	}
	g.p.SetPrompt(strings.Join(args, " "))
	g.p.Write("Prompt set.")
	return nil
}

// DoColor will toggle the color flag for a player.
func (g *Game) DoColor(ctx context.Context, args ...string) error {
	if v := g.p.ToggleFlag("color"); v {
		g.p.Write("{gColor enabled!{x")
	} else {
		g.p.Write("Color disabled :(")
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
			p.DisableFlag("automap")
			p.Write("Automap turned off.")
			return nil
		case args[0] == "on":
			p.EnableFlag("automap")
			p.Write("Automap turned on.")
			return nil
		default:
			r, err := strconv.Atoi(args[0])
			if err != nil {
				p.Write("You must specify a number for your map size, i.e. 'map 5'")
			}
			radius = int64(r)
		}
	}
	if radius > 100 || radius == 0 {
		radius = int64(100)
	}
	g.p.Write(p.Map(radius))
	return nil
}
