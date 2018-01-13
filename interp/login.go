package interp

import (
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"io"
	"io/ioutil"

	"github.com/Cidan/gomud/player"
	"github.com/Cidan/gomud/state"
)

// Login interp for handling user login
type Login struct {
	p     *player.Player
	state *state.State
}

func NewLogin(p *player.Player) *Login {
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

func hashPassword(pw string) string {
	h := sha512.New()
	io.WriteString(h, pw)
	return hex.EncodeToString(h.Sum(nil))
}

func (l *Login) Read(text string) error {
	return l.state.Process(text)
}

func (l *Login) AskName(text string) error {
	// Check for save
	// TODO: Validate name
	// TODO: this is extremely unsafe.
	data, err := ioutil.ReadFile("/tmp/" + text)
	if err != nil {
		l.p.Write("Are you sure you want to be known as %s?", text)
		l.p.Data.Name = text
		return l.state.SetState("CONFIRM_NAME")
	}
	err = json.Unmarshal(data, &l.p.Data)
	if err != nil {
		return err
	}
	l.p.Write("Password:")
	return l.state.SetState("ASK_PASSWORD")
}

func (l *Login) AskPassword(text string) error {
	if l.p.Data.Password != hashPassword(text) {
		l.p.Write("Wrong password. Bye.")
		l.p.Stop()
		return nil
	}
	l.p.Write("Entering the world!")
	l.p.Interp = NewGame(l.p)
	return nil
}

func (l *Login) ConfirmName(text string) error {
	if text != "yes" && text != "y" {
		l.p.Write("Okay, so what's your name?")
		return l.state.SetState("ASK_NAME")
	}

	l.p.Write("Welcome %s, please give me a password.", l.p.Data.Name)
	return l.state.SetState("NEW_PASSWORD")
}

func (l *Login) NewPassword(text string) error {
	// TODO: validate password
	pw := hashPassword(text)
	l.p.Data.Password = pw
	l.p.Write("Confirm your password and type it again")
	return l.state.SetState("CONFIRM_PASSWORD")
}

func (l *Login) ConfirmPassword(text string) error {
	pw := hashPassword(text)
	if pw != l.p.Data.Password {
		l.p.Write("Passwords do not match\n")
		l.p.Write("Let's, try this again. Please give me a new password.")
		return l.state.SetState("NEW_PASSWORD")
	}
	l.p.Write("Entering the world!")
	l.p.Interp = NewGame(l.p)

	return nil
}
