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

func LogPlayerIn(t *testing.T, reader *bufio.Reader, writer *bufio.Writer) {
	loginCommands := []string{
		"Test",
		"yes",
		"pass",
		"pass",
		"build",
		"dig east",
	}

	// Read the login text first.
	_, err := reader.ReadString('\r')
	assert.Nil(t, err)
	for _, command := range loginCommands {
		writer.WriteString(command + "\n")
		writer.Flush()
		repl, err := reader.ReadString('\r')
		assert.Nil(t, err)
		fmt.Printf("%s\n", repl)
	}
}

// This test rapidly moves a player between two rooms, while deleting a room in
// another go routine, ensuring room deletes are concurrently safe.
func TestPlayerMovementRace(t *testing.T) {
	config.Set("save_path", t.TempDir())
	makeStartingRoom()
	p := NewPlayer()
	assert.NotNil(t, p)
	server := server.New()
	go server.Listen(2000, func(c net.Conn) {
		// Simulated player connection loop
		p.SetConnection(c)
		go p.Start()
	})

	conn, err := net.Dial("tcp", "localhost:2000")
	assert.Nil(t, err)
	reader := bufio.NewReader(conn)
	writer := bufio.NewWriter(conn)

	LogPlayerIn(t, reader, writer)

	c := make(chan bool)
	go func() {
		for i := 0; i < 100; i++ {
			writer.WriteString("dig east\n")
			assert.Nil(t, writer.Flush())
			repl, _ := reader.ReadString('\r')
			fmt.Printf("%s for reply\n", repl)
			writer.WriteString("west\n")
			assert.Nil(t, writer.Flush())

			reader.ReadString('\r')
		}
		c <- true
	}()

	for i := 0; i < 10000; i++ {
		r := Atlas.GetRoom(1, 0, 0)
		if r != nil {
			r.Delete()
		}
	}
	<-c
}
