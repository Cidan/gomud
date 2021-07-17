package construct

import (
	"context"
	"strings"

	"github.com/Cidan/gomud/lock"
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
	quit     bool
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

func (e *TextInterp) Read(ctx context.Context, text string) error {
	all := strings.SplitN(text, " ", 2)
	if e.commands.Has(all[0]) {
		return e.commands.Process(ctx, all[0], all[1:]...)
	}
	e.quit = false
	e.buffer += text + "\n"
	e.p.Write(ctx, e.buffer)
	return nil
}

func (e *TextInterp) Start(ctx context.Context, field *string) context.Context {
	e.buffer = ""
	ectx, cancel := context.WithCancel(lock.Context(ctx, e.p.GetData(ctx).UUID+"text_edit"))
	e.context = ectx
	e.cancel = cancel
	e.field = field
	e.quit = false
	return e.context
}

// Text returns the edited buffer.
func (e *TextInterp) Text() string {
	return e.buffer
}

func (e *TextInterp) DoDone(ctx context.Context, args ...string) error {
	result := strings.TrimSuffix(e.buffer, "\n")
	*e.field = result
	e.field = nil
	e.p.Write(ctx, "{GText saved.{x")
	e.cancel()
	return nil
}

func (e *TextInterp) DoCancel(ctx context.Context, args ...string) error {
	if e.quit {
		e.field = nil
		e.p.Write(ctx, "{RCancelling editing, text not saved.{x")
		e.cancel()
	} else {
		e.p.Write(ctx, "Type :q to quit again. Any other command will back out.")
		e.quit = true
	}
	return nil
}
