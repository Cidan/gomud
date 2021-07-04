package construct

import (
	"github.com/Cidan/gomud/state"
)

// Login interp for handling user login
type Login struct {
	p     *Player
	state *state.State
}

// NewLogin interp, to handle user login and character creation.
func NewLoginInterp(p *Player) *Login {
	l := &Login{p: p}

	// Create our state flow
	s := state.New("ASK_NAME")
	s.
		Add(&state.Event{
			Name: "ASK_NAME",
			Fn:   l.AskName,
		}).
		Add(&state.Event{
			Name: "ASK_PASSWORD",
			Fn:   l.AskPassword,
		}).
		Add(&state.Event{
			Name: "CONFIRM_NAME",
			Fn:   l.ConfirmName,
		}).
		Add(&state.Event{
			Name: "NEW_PASSWORD",
			Fn:   l.NewPassword,
		}).
		Add(&state.Event{
			Name: "CONFIRM_PASSWORD",
			Fn:   l.ConfirmPassword,
		})
	l.state = s
	return l
}

func (l *Login) Read(text string) error {
	return l.state.Process(text)
}

// AskName step.
func (l *Login) AskName(text string) error {
	// Check for save
	// TODO: Validate name
	// TODO: this is extremely unsafe.
	l.p.SetName(text)
	loaded, err := l.p.Load()
	if err == nil && !loaded {
		l.p.Write("Are you sure you want to be known as %s?", text)
		return l.state.SetState("CONFIRM_NAME")
	}
	if err != nil {
		l.p.Write("Something went wrong trying to load your pfile.")
		l.p.Stop()
	}
	l.p.Write("Password: ")
	return l.state.SetState("ASK_PASSWORD")
}

// AskPassword step.
func (l *Login) AskPassword(text string) error {
	if !l.p.IsPassword(text) {
		l.p.Write("Wrong password. Bye.")
		l.p.Stop()
		return nil
	}
	l.p.Write("Entering the world!")

	if target := GetRoomByUUID(l.p.Data.Room); target != nil {
		l.p.ToRoom(target)
	} else {
		l.p.ToRoom(GetRoom(0, 0, 0))
	}

	l.p.Game()
	l.p.Command("look")
	return nil
}

// ConfirmName step.
func (l *Login) ConfirmName(text string) error {
	if text != "yes" && text != "y" {
		l.p.Write("Okay, so what's your name?")
		return l.state.SetState("ASK_NAME")
	}

	l.p.Write("Welcome %s, please give me a password: ", l.p.GetName())
	return l.state.SetState("NEW_PASSWORD")
}

// NewPassword step.
func (l *Login) NewPassword(text string) error {
	// TODO: validate password
	l.p.SetPassword(text)
	l.p.Write("Confirm your password and type it again: ")
	return l.state.SetState("CONFIRM_PASSWORD")
}

// ConfirmPassword step.
func (l *Login) ConfirmPassword(text string) error {
	if !l.p.IsPassword(text) {
		l.p.Write("Passwords do not match\n")
		l.p.Write("Let's try this again. Please give me a new password: ")
		return l.state.SetState("NEW_PASSWORD")
	}
	l.p.Write("Entering the world!")
	l.p.Game()
	l.p.ToRoom(GetRoom(0, 0, 0))
	l.p.Command("look")
	return nil
}
