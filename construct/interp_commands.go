package construct

import (
	"context"
	"fmt"
)

type commandCallback func(context.Context, ...string) error

// Interp type for interpreting player commands.
type Interp interface {
	Read(context.Context, string) error
}

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

var (
	//ErrCommandNotFound denotes when a command does not exist for an interp.
	ErrCommandNotFound = fmt.Errorf("command does not exist")
)

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

func (c *commandMap) Process(ctx context.Context, command string, input ...string) error {
	if c.commands[command] != nil {
		return c.commands[command].Fn(ctx, input...)
	}

	return ErrCommandNotFound
}

func (c *commandMap) Has(command string) bool {
	_, ok := c.commands[command]
	return ok
}
