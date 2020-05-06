package interp

import "github.com/Cidan/gomud/types"

type Build struct {
	p types.Player
}

func NewBuild(p types.Player) *Build {
	b := &Build{
		p: p,
	}
	return b
}

func (b *Build) Read(text string) error {
	return nil
}
