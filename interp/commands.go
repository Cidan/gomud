package interp

type commandCallback func(player, ...string) error

type command struct {
	name  string
	alias []string
	state []string
	Fn    commandCallback
}

// CommandMap is the top level object for commands
type commandMap struct {
	commands map[string]*command
}

func newCommands() *commandMap {
	return &commandMap{
		commands: make(map[string]*command),
	}
}

func (c *commandMap) Add(nc *command) *commandMap {
	if nc.name == "" {
		panic("Command added with a blank name. Fix this.")
	}

	c.commands[nc.name] = nc
	for _, alias := range nc.alias {
		c.commands[alias] = nc
	}
	return c
}

func (c *commandMap) Process(p player, command string, input ...string) error {
	if c.commands[command] != nil {
		return c.commands[command].Fn(p, input...)
	}

	p.Write("Huh?")
	return nil
}
