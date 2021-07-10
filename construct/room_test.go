package construct

import (
	"bufio"
	"net"
	"testing"

	"github.com/Cidan/gomud/config"
	"github.com/Cidan/gomud/mocks/server"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
)

func LogPlayerIn(t *testing.T, reader *bufio.Reader, writer *bufio.Writer) {
	loginCommands := []string{
		"Test",
		"yes",
		"pass",
		"pass",
		"build",
	}

	// Read the login text first.
	_, err := reader.ReadString('\r')
	assert.Nil(t, err)
	for _, command := range loginCommands {
		writer.WriteString(command + "\n")
		writer.Flush()
		_, err := reader.ReadString('\r')
		assert.Nil(t, err)
	}
}

// This test rapidly moves a player between two rooms, while deleting a room in
// another go routine, ensuring room deletes are concurrently safe.
// This test should be run with race detection via `go test -race -count=3`
func TestPlayerMovementRace(t *testing.T) {
	log.Level(zerolog.Disabled)
	config.Set("save_path", t.TempDir())
	makeStartingRoom()
	p := NewPlayer()
	assert.NotNil(t, p)
	server := server.New()
	go server.Listen(2020, func(c net.Conn) {
		// Simulated player connection loop
		p.SetConnection(c)
		go p.Start()
	})

	conn, err := net.Dial("tcp", "localhost:2020")
	assert.Nil(t, err)
	reader := bufio.NewReader(conn)
	writer := bufio.NewWriter(conn)

	LogPlayerIn(t, reader, writer)

	c := make(chan bool)
	go func(reader *bufio.Reader, writer *bufio.Writer) {
		for i := 0; i < 100; i++ {
			writer.WriteString("dig east\n")
			assert.Nil(t, writer.Flush())
			reader.ReadString('\r')
			writer.WriteString("west\n")
			assert.Nil(t, writer.Flush())

			reader.ReadString('\r')
		}
		c <- true
	}(reader, writer)

	for i := 0; i < 10000; i++ {
		r := Atlas.GetRoom(1, 0, 0)
		if r != nil {
			r.Delete()
		}
	}
	<-c
	server.Close()
}

func TestEditRoom(t *testing.T) {
	config.Set("save_path", t.TempDir())
	makeStartingRoom()
	p := NewPlayer()
	assert.NotNil(t, p)
	server := server.New()
	go server.Listen(2021, func(c net.Conn) {
		// Simulated player connection loop
		p.SetConnection(c)
		go p.Start()
	})

	conn, err := net.Dial("tcp", "localhost:2021")
	assert.Nil(t, err)
	reader := bufio.NewReader(conn)
	writer := bufio.NewWriter(conn)

	LogPlayerIn(t, reader, writer)

	writer.WriteString("edit room description\n")
	assert.Nil(t, writer.Flush())
	reader.ReadString('\r')

	writer.WriteString(":w\n")
	assert.Nil(t, writer.Flush())
	reader.ReadString('\r')

	server.Close()
}
