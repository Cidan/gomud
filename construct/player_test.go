package construct

import (
	"bufio"
	"fmt"
	"net"
	"testing"

	"github.com/Cidan/gomud/config"
	"github.com/Cidan/gomud/mocks/server"
	"github.com/stretchr/testify/assert"
)

func setupTest(t *testing.T, port int) (*bufio.Reader, *bufio.Writer, *server.Server) {
	t.Helper()
	config.Set("save_path", t.TempDir())
	makeStartingRoom()
	p := NewPlayer()
	assert.NotNil(t, p)
	server := server.New()
	go server.Listen(port, func(c net.Conn) {
		// Simulated player connection loop
		p.SetConnection(c)
		go p.Start()
	})

	conn, err := net.Dial("tcp", fmt.Sprintf("localhost:%d", port))
	assert.Nil(t, err)
	reader := bufio.NewReader(conn)
	writer := bufio.NewWriter(conn)
	return reader, writer, server
}

func runCommands(t *testing.T, r *bufio.Reader, w *bufio.Writer, commands []string) {
	for _, command := range commands {
		w.WriteString(command + "\n")
		assert.Nil(t, w.Flush())
		_, err := r.ReadString('\r')
		assert.Nil(t, err)
	}
}

func TestPlayerMap(t *testing.T) {
	r, w, s := setupTest(t, 2023)
	LogPlayerIn(t, r, w)
	var commands = []string{
		"dig west",
		"map 10",
	}

	runCommands(t, r, w, commands)
	s.Close()
}

func TestPlayerSaveAndQuit(t *testing.T) {
	r, w, s := setupTest(t, 2024)
	LogPlayerIn(t, r, w)
	var commands = []string{
		"save",
	}

	runCommands(t, r, w, commands)
	s.Close()
}

func TestPlayerInterpSwap(t *testing.T) {
	r, w, s := setupTest(t, 2025)
	LogPlayerIn(t, r, w)
	var commands = []string{
		"build",
		"build",
		"edit room name",
		"This is a name",
		":w",
		"save",
	}

	runCommands(t, r, w, commands)
	s.Close()
}
