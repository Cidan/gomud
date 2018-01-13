package interp

import "github.com/Cidan/gomud/player"

type CommandCallback func(*player.Player, ...string) error

type Command struct {
	name  string
	alias []string
	state []string
	Fn    CommandCallback
}

// Commands is the top level object for commands
type CommandMap struct {
	commands map[string]*Command
}

func NewCommands() *CommandMap {
	return &CommandMap{
		commands: make(map[string]*Command),
	}
}

func (c *CommandMap) Add(nc *Command) *CommandMap {
	if nc.name == "" {
		panic("Command added with a blank name. Fix this.")
	}

	c.commands[nc.name] = nc
	for _, alias := range nc.alias {
		c.commands[alias] = nc
	}
	return c
}

func (c *CommandMap) Process(p *player.Player, command string, input ...string) error {
	if c.commands[command] != nil {
		return c.commands[command].Fn(p, input...)
	}

	p.Write("Huh?")
	return nil
}
