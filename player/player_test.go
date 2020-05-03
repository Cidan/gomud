package player

import (
	"bufio"
	"net"
	"strings"
	"sync"
	"testing"

	"github.com/Cidan/gomud/mocks/server"
	"github.com/stretchr/testify/assert"
)

var testCommands = []string{
	"buffer test",
	"name",
}

func TestPlayer(t *testing.T) {
	wg := sync.WaitGroup{}
	wg.Add(1)

	p := New()
	assert.NotNil(t, p)
	server := server.New()
	go server.Listen(2000, func(c net.Conn) {
		p.SetConnection(c)
		p.Buffer("buffer %s\n", "test")
		p.Flush()
		p.SetName("name")
		p.Write(p.GetName() + "\n")
		err := p.Save()
		assert.Nil(t, err)
		loaded, err := p.Load()
		assert.Nil(t, err)
		assert.True(t, loaded)

		p.SetPassword("password")
		assert.True(t, p.IsPassword("password"))
		p.Stop()
		wg.Done()
	})

	conn, err := net.Dial("tcp", "localhost:2000")
	assert.Nil(t, err)
	reader := bufio.NewReader(conn)

	for _, testCase := range testCommands {
		text, err := reader.ReadString('\n')
		text = strings.TrimSpace(text)
		assert.Nil(t, err)
		assert.Equal(t, testCase, text)
	}
	wg.Wait()
}
