package state

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEndToEnd(t *testing.T) {
	s := New("START")

	s.Add(&Event{
		Name: "START",
		Fn: func(str string) error {
			assert.Equal(t, "help", str)
			return nil
		},
	}).Add(&Event{
		Name: "SET_OUTSIDE",
		Fn: func(str string) error {
			assert.Equal(t, "set outside", str)
			s.SetState("SET_INSIDE")
			return nil
		},
	}).Add(&Event{
		Name: "SET_INSIDE",
		Fn: func(str string) error {
			assert.Equal(t, "set inside", str)
			return nil
		},
	})

	assert.Nil(t, s.Process("help"))
	s.SetState("SET_OUTSIDE")
	assert.Nil(t, s.Process("set outside"))
	assert.Nil(t, s.Process("set inside"))
	assert.Error(t, s.SetState("invalid_state"))
}
