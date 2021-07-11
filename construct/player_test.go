package construct

import (
	"bufio"
	"testing"

	"github.com/stretchr/testify/assert"
)

func runCommands(t *testing.T, r *bufio.Reader, w *bufio.Writer, commands []string) {
	for _, command := range commands {
		w.WriteString(command + "\n")
		assert.NoError(t, w.Flush())
		_, err := r.ReadString('\r')
		assert.NoError(t, err)
	}
}

func TestPlayerMap(t *testing.T) {
	s := testSetupServer(t, 2023)
	r, w := testLoginNewUser(t, "PlayerMap", s)
	var commands = []string{
		"dig west",
		"map 10",
	}

	runCommands(t, r, w, commands)
	s.Close()
}

func TestPlayerSaveAndQuit(t *testing.T) {
	s := testSetupServer(t, 2023)
	r, w := testLoginNewUser(t, "SaveAndQuit", s)
	var commands = []string{
		"save",
	}

	runCommands(t, r, w, commands)
	s.Close()
}

func TestPlayerInterpSwap(t *testing.T) {
	s := testSetupServer(t, 2023)
	r, w := testLoginNewUser(t, "InterSwap", s)
	var commands = []string{
		"build",
		"build",
		"edit room name",
		"This is a name",
		":q",
		":q",
		"edit room name",
		"This is a name",
		":w",
		"save",
	}

	runCommands(t, r, w, commands)
	s.Close()
}
