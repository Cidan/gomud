package interp

import (
	"strings"

	"github.com/Cidan/gomud/types"
	"github.com/rs/zerolog/log"
)

type Build struct {
	p types.Player
}

var buildCommands *commandMap

func init() {
	buildCommands = newCommands()
	buildCommands.Add(&command{
		name:  "look",
		alias: []string{"l"},
		Fn:    DoLook,
	})
}

func NewBuild(p types.Player) *Build {
	b := &Build{
		p: p,
	}
	return b
}

func (b *Build) Read(text string) error {
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

	return gameCommands.Process(b.p, all[0], all[1:]...)
}
