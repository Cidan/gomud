package state

import "errors"
import "fmt"

type EventCallback func(string) error

type Event struct {
	Name string
	Fn   EventCallback
}

// State machine for handling work flows.
type State struct {
	current string
	states  map[string]*Event
}

func New(initial string) *State {
	return &State{
		states:  make(map[string]*Event),
		current: initial,
	}
}

func (s *State) Add(e *Event) *State {
	s.states[e.Name] = e
	return s
}

func (s *State) SetState(name string) error {
	if s.states[name] == nil {
		return fmt.Errorf("No such state defined, %s", name)
	}
	s.current = name
	return nil
}

func (s *State) Process(text string) error {
	if s.states[s.current] == nil {
		return errors.New("Invalid state set")
	}
	return s.states[s.current].Fn(text)
}
