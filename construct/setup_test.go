package construct

import (
	"bufio"
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/Cidan/gomud/config"
	"github.com/Cidan/gomud/mocks/server"
	"github.com/stretchr/testify/assert"
)

func testSetupServer(t *testing.T, port int) *server.Server {
	t.Helper()
	config.Set("save_path", t.TempDir())
	makeStartingRoom()
	server := server.New(port)
	go server.Listen(func(c net.Conn) {
		// Simulated player connection loop
		p := NewPlayer()
		assert.NotNil(t, p)
		p.SetConnection(c)
		go p.Start()
	})
	time.Sleep(time.Millisecond * 500)
	return server
}

func testLoginNewUser(t *testing.T, name string, server *server.Server) (*bufio.Reader, *bufio.Writer) {
	t.Helper()

	conn, err := net.Dial("tcp", fmt.Sprintf("localhost:%d", server.Port))
	assert.NoError(t, err)
	reader := bufio.NewReader(conn)
	writer := bufio.NewWriter(conn)

	loginCommands := []string{
		name,
		"yes",
		"pass",
		"pass",
		"build",
	}

	// Read the login text first.
	recv, err := reader.ReadString('\r')
	fmt.Printf("testLoginNewUser(): got %s\n", recv)
	assert.NoError(t, err)
	for _, command := range loginCommands {
		writer.WriteString(command + "\n")
		writer.Flush()
		recv, err := reader.ReadString('\r')
		fmt.Printf("testLoginNewUser(): command got: %s\n", recv)
		assert.NoError(t, err)
	}

	return reader, writer
}

func testLoginUser(t *testing.T, name string, server *server.Server) (*bufio.Reader, *bufio.Writer) {
	t.Helper()

	conn, err := net.Dial("tcp", fmt.Sprintf("localhost:%d", server.Port))
	assert.NoError(t, err)
	reader := bufio.NewReader(conn)
	writer := bufio.NewWriter(conn)

	loginCommands := []string{
		name,
		"pass",
		"build",
	}

	// Read the login text first.
	recv, err := reader.ReadString('\r')
	fmt.Printf("testLoginUser(): got %s\n", recv)
	assert.NoError(t, err)
	for _, command := range loginCommands {
		writer.WriteString(command + "\n")
		writer.Flush()
		recv, err := reader.ReadString('\r')
		fmt.Printf("testLoginUser(): command got: %s\n", recv)
		assert.NoError(t, err)
	}

	return reader, writer
}
