package construct

import (
	"bufio"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func runCommands(t *testing.T, r *bufio.Reader, w *bufio.Writer, commands []string) {
	/*
		go func() {
			for {
				r.ReadString('\xfa')
			}
		}()
	*/
	for _, command := range commands {
		w.WriteString(command + "\n")
		assert.NoError(t, w.Flush())
		/*
			for {
				recv, err := r.ReadString('\xf9')
				assert.NoError(t, err)
				if strings.Contains(recv, "OKPROMPT") {
					break
				}
			}
		*/
		//		recv, err := r.ReadString('\r')
		//		fmt.Printf("got %s\n", recv)
		//		assert.NoError(t, err)
	}
}

func TestPlayerMap(t *testing.T) {
	testSetupWorld(t)
	r, w := testLoginNewUser(t, "PlayerMap")
	var commands = []string{
		"dig west",
		"map 10",
	}

	runCommands(t, r, w, commands)
}

func TestPlayerSaveAndQuit(t *testing.T) {
	testSetupWorld(t)
	r, w := testLoginNewUser(t, "SaveAndQuit")
	var commands = []string{
		"save",
	}

	runCommands(t, r, w, commands)
}

func TestPlayerInterpSwap(t *testing.T) {
	testSetupWorld(t)
	r, w := testLoginNewUser(t, "InterSwap")
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
}

func TestPlayerReconnect(t *testing.T) {
	testSetupWorld(t)
	r, w := testLoginNewUser(t, "PlayerReconnect")
	fmt.Printf("on to the real test\n")
	runCommands(t, r, w, []string{
		"save",
	})
	fmt.Printf("About to try a reconnect\n")
	rl, wl := testLoginUser(t, "PlayerReconnect")
	runCommands(t, rl, wl, []string{
		"say hi",
		"save",
		"quit",
	})
}
