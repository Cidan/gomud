package construct

import (
	"bufio"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func runCommands(t *testing.T, r *bufio.Reader, w *bufio.Writer, commands []string) {
	for _, command := range commands {
		w.WriteString(command + "\n")
		assert.NoError(t, w.Flush())
		recv, err := r.ReadString('\r')
		fmt.Printf("got %s\n", recv)
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

func TestPlayerReconnect(t *testing.T) {
	s := testSetupServer(t, 2024)
	r, w := testLoginNewUser(t, "PlayerReconnect", s)
	runCommands(t, r, w, []string{
		"save",
		"save",
		"save",
		"save",
	})
	fmt.Printf("About to try a reconnect\n")
	rl, wl := testLoginUser(t, "PlayerReconnect", s)
	runCommands(t, rl, wl, []string{
		"say hi",
		"save",
		"quit",
	})
	s.Close()
	time.Sleep(5 * time.Second)
}
