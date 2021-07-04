package construct

import (
	"bufio"
	"net"
	"testing"

	"github.com/Cidan/gomud/mocks/server"
	"github.com/stretchr/testify/assert"
)

type testCase struct {
	name     string
	input    []string
	response []string
}

// Insert your tests cases here in the form of text sent : text response.
// Do not insert newlines, unless you specifically are testing for newlines.
var testCases = []testCase{
	{
		"Login Name",
		[]string{"name"},
		[]string{"Are you sure you want to be known as name?"},
	},
	{
		"Confirm Name",
		[]string{"yes"},
		[]string{"Welcome name, please give me a password: "},
	},
	{
		"New Password",
		[]string{"pass123"},
		[]string{"Confirm your password and type it again: "},
	},
	{
		"Fail Confirm Password",
		[]string{"wrongpass"},
		[]string{"Passwords do not match\n", "Let's try this again. Please give me a new password: "},
	},
	{
		"New Password Again",
		[]string{"pass123"},
		[]string{"Confirm your password and type it again: "},
	},
	// TODO(lobato): This crashes tests right now -- no game world loaded.
	/*
		{
			"Confirm Password",
			[]string{"pass123"},
			[]string{"Entering the world!"},
		},
	*/
}

func TestPlayer(t *testing.T) {

	p := NewPlayer()
	assert.NotNil(t, p)
	server := server.New()
	go server.Listen(2000, func(c net.Conn) {
		// Simulated player connection loop
		p.SetConnection(c)
		p.Start()
	})

	conn, err := net.Dial("tcp", "localhost:2000")
	assert.Nil(t, err)
	reader := bufio.NewReader(conn)
	writer := bufio.NewWriter(conn)

	// Read the login text first.
	loginText, err := reader.ReadString('\r')
	assert.Nil(t, err)
	assert.Equal(t, "Welcome, by what name are you known?\r", loginText)

	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			for _, input := range test.input {
				_, err := writer.WriteString(input + "\n")
				assert.Nil(t, err)
				err = writer.Flush()
				assert.Nil(t, err)
			}

			for _, response := range test.response {
				recv, err := reader.ReadString('\r')
				assert.Nil(t, err)
				assert.Equal(t, response+"\r", recv)
			}
		})
	}

}
