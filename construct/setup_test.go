package construct

import (
	"bufio"
	"fmt"
	"net"
	"testing"

	"github.com/Cidan/gomud/config"
	"github.com/stretchr/testify/assert"
)

func testSetupWorld(t *testing.T) {
	t.Helper()
	config.Set("save_path", t.TempDir())
	makeStartingRoom()
}

func testLoginNewUser(t *testing.T, name string) (*bufio.Reader, *bufio.Writer) {
	t.Helper()

	client, server := net.Pipe()
	p := NewPlayer()
	p.SetConnection(server)
	go p.Start()

	reader := bufio.NewReader(client)
	writer := bufio.NewWriter(client)

	loginCommands := []string{
		name,
		"yes",
		"pass",
		"pass",
	}

	// Read the login text first.
	recv, err := reader.ReadString('\xf9')
	fmt.Printf("testLoginNewUser(): got %s\n", recv)
	assert.NoError(t, err)
	for _, command := range loginCommands {
		fmt.Printf("testLoginNewUser(): command sent: %s\n", command)
		writer.WriteString(command + "\n")
		writer.Flush()
		recv, err := reader.ReadString('\xf9')
		fmt.Printf("testLoginNewUser(): command got: %s\n", recv)
		assert.NoError(t, err)
	}

	return reader, writer
}

func testLoginUser(t *testing.T, name string) (*bufio.Reader, *bufio.Writer) {
	t.Helper()

	client, server := net.Pipe()
	p := NewPlayer()
	p.SetConnection(server)
	go p.Start()

	reader := bufio.NewReader(client)
	writer := bufio.NewWriter(client)

	loginCommands := []string{
		name,
		"pass",
	}

	// Read the login text first.
	recv, err := reader.ReadString('\xf9')
	fmt.Printf("testLoginUser(): got %s\n", recv)
	assert.NoError(t, err)
	for _, command := range loginCommands {
		writer.WriteString(command + "\n")
		writer.Flush()
		recv, err := reader.ReadString('\xf9')
		fmt.Printf("testLoginUser(): command got: %s\n", recv)
		assert.NoError(t, err)
	}

	return reader, writer
}
