package construct

import (
	"context"
	"strings"
)

// TextInterp is the editor interp, which allows for long form
// multi-line editing from a MUD client. This is useful for editing
// item, room, and object descriptions, writing help files, long
// form lore pieces, etc. The editor is a WYSIWYG editor that accurately
// allows for previews as a player would see the text in game.
type TextInterp struct {
	p        *Player
	commands *commandMap
	context  context.Context
	cancel   context.CancelFunc
	buffer   string
	field    *string
}

func NewTextInterp(p *Player) *TextInterp {
	e := &TextInterp{
		p: p,
	}
	commands := newCommands()
	commands.Add(&command{
		name: `:w`,
		Fn:   e.DoDone,
	}).Add(&command{
		name: `:q`,
		Fn:   e.DoCancel,
	})
	e.commands = commands
	return e
}

func (e *TextInterp) Read(text string) error {
	all := strings.SplitN(text, " ", 2)
	if e.commands.Has(all[0]) {
		return e.commands.Process(all[0], all[1:]...)
	}
	e.buffer += text + "\n"
	e.p.Write(e.buffer)
	return nil
}

func (e *TextInterp) Start(field *string) context.Context {
	e.buffer = ""
	ctx, cancel := context.WithCancel(context.Background())
	e.context = ctx
	e.cancel = cancel
	e.field = field
	return e.context
}

// Text returns the edited buffer.
func (e *TextInterp) Text() string {
	return e.buffer
}

func (e *TextInterp) DoDone(args ...string) error {
	result := strings.TrimSuffix(e.buffer, "\n")
	*e.field = result
	e.field = nil
	e.p.Write("{GText saved.{x")
	e.cancel()
	return nil
}

func (e *TextInterp) DoCancel(args ...string) error {
	e.field = nil
	e.p.Write("{RCancelling editing, text not saved.{x")
	e.cancel()
	return nil
}
