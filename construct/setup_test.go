package construct

import (
	"bufio"
	"fmt"
	"net"
	"testing"

	"github.com/Cidan/gomud/config"
	"github.com/Cidan/gomud/lock"
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
	ctx := lock.Context(p.Context(), p.Data.UUID+"incomming_conn")
	p.SetConnection(ctx, server)
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
	go func() {
		for {
			recv, _ := reader.ReadString('\xf9')
			fmt.Printf("testLoginNewUser(): got %s\n", recv)
		}
	}()
	/*
		recv, err := reader.ReadString('\xf9')
		fmt.Printf("testLoginNewUser(): got %s\n", recv)
		assert.NoError(t, err)
	*/
	for _, command := range loginCommands {
		fmt.Printf("testLoginNewUser(): command sent: %s\n", command)
		writer.WriteString(command + "\n")
		writer.Flush()
		/*
			recv, err := reader.ReadString('\xf9')
			fmt.Printf("testLoginNewUser(): command got: %s\n", recv)
			assert.NoError(t, err)
		*/
	}

	return reader, writer
}

func testLoginUser(t *testing.T, name string) (*bufio.Reader, *bufio.Writer) {
	t.Helper()

	client, server := net.Pipe()
	p := NewPlayer()
	ctx := lock.Context(p.Context(), p.Data.UUID+"incomming_conn")
	p.SetConnection(ctx, server)
	go p.Start()

	reader := bufio.NewReader(client)
	writer := bufio.NewWriter(client)

	loginCommands := []string{
		name,
		"pass",
	}
	go func() {
		for {
			reader.ReadString('\xf9')
		}
	}()
	// Read the login text first.
	/*
		recv, err := reader.ReadString('\xf9')
		fmt.Printf("testLoginUser(): got %s\n", recv)
		assert.NoError(t, err)
	*/
	for _, command := range loginCommands {
		writer.WriteString(command + "\n")
		writer.Flush()
		/*
			recv, err := reader.ReadString('\xf9')
			fmt.Printf("testLoginUser(): command got: %s\n", recv)
			assert.NoError(t, err)
		*/
	}

	return reader, writer
}
