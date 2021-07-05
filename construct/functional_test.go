package construct

// This test file contains functional tests for the game world.
// TODO(lobato): Move this out of construct and into the execution endpoint
// to simulate true functional tests.
import (
	"bufio"
	"net"
	"os"
	"testing"

	"github.com/Cidan/gomud/color"
	"github.com/Cidan/gomud/config"
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
		"Back Out Login Name",

		[]string{"no"},
		[]string{"Okay, so what's your name?"},
	},
	{
		"Actual Login Name",

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

	{
		"Confirm Password",

		[]string{"pass123"},
		[]string{"Entering the world!"},
	},
	{
		"First Look",
		[]string{},
		[]string{"\n\nThe Alpha\n\n  It all starts here.\n"},
	},

	// Begin post game test cases -- add loaded world/player cases below this line.
	// Build test commands
	{
		"Build",
		[]string{"build"},
		[]string{"Entering build mode\n"},
	},
	{
		"Dig",
		[]string{"dig"},
		[]string{"Which direction do you want to dig?"},
	},
	{
		"Dig North",
		[]string{"dig north"},
		[]string{"\n\nNew Room\n\n  This is a new room, with a new description.\n"},
	},
	{
		"Dig South Already Exists",
		[]string{"dig south"},
		[]string{"There's already a room 'south'.\n"},
	},
	{
		"Dig West",
		[]string{"dig west"},
		[]string{"\n\nNew Room\n\n  This is a new room, with a new description.\n"},
	},
	{
		"Dig South",
		[]string{"dig south"},
		[]string{"\n\nNew Room\n\n  This is a new room, with a new description.\n"},
	},
	{
		"Dig Up",
		[]string{"dig up"},
		[]string{"\n\nNew Room\n\n  This is a new room, with a new description.\n"},
	},
	{
		"Dig East",
		[]string{"dig east"},
		[]string{"\n\nNew Room\n\n  This is a new room, with a new description.\n"},
	},
	{
		"Dig East Again",
		[]string{"dig east"},
		[]string{"\n\nNew Room\n\n  This is a new room, with a new description.\n"},
	},
	{
		"Dig Down",
		[]string{"dig down"},
		[]string{"\n\nNew Room\n\n  This is a new room, with a new description.\n"},
	},
	{
		"Dig Nonsense",
		[]string{"dig sdjlkasdlj"},
		[]string{"That's not a valid direction to dig in."},
	},
	{
		"Autobuild On",
		[]string{"autobuild"},
		[]string{"Autobuild has been enabled."},
	},
	{
		"Autobuild Off",
		[]string{"autobuild"},
		[]string{"Autobuild has been disabled."},
	},
	{
		"Build Off",
		[]string{"build"},
		[]string{"Build mode deactivated."},
	},
	// Build tests done.
	{
		"Go North",
		[]string{"north"},
		[]string{"You can't go that way!"},
	},
	{
		"Go South",
		[]string{"south"},
		[]string{"You can't go that way!"},
	},
	{
		"Go East",
		[]string{"east"},
		[]string{"You can't go that way!"},
	},
	{
		"Go West",
		[]string{"west"},
		[]string{"\n\nThe Alpha\n\n  It all starts here.\n"},
	},
	{
		"Go Up",
		[]string{"up"},
		[]string{"\n\nNew Room\n\n  This is a new room, with a new description.\n"},
	},
	{
		"Go Down",
		[]string{"down"},
		[]string{"\n\nThe Alpha\n\n  It all starts here.\n"},
	},
	{
		"Nonsense Command",
		[]string{"asokdjasljdk"},
		[]string{"Huh?"},
	},
	{
		"Save",
		[]string{"save"},
		[]string{"Your player has been saved."},
	},
	{
		"Disable Prompt",
		[]string{"prompt"},
		[]string{"Prompt disabled."},
	},
	{
		"Enable Prompt",
		[]string{"prompt"},
		[]string{"Prompt enabled."},
	},
	{
		"Set Prompt",
		[]string{"prompt <%h>"},
		[]string{"Prompt set."},
	},
	{
		"Disable Color",
		[]string{"color"},
		[]string{"Color disabled :("},
	},
	{
		"Enable Color",
		[]string{"color"},
		[]string{"{gColor enabled!{x"},
	},
	{
		"Say Nothing",
		[]string{"say"},
		[]string{"Say what?"},
	},
	{
		"Say hi",
		[]string{"say hi"},
		[]string{"{yYou say, {x'hi{x'"},
	},
	// This should always be the last test. Do not change this case.
	{
		"Quit",
		[]string{"quit"},
		[]string{"See ya!\n"},
	},
}

func makeStartingRoom() {
	room := NewRoom(&RoomData{
		Name:        "The Alpha",
		Description: "It all starts here.",
		X:           0,
		Y:           0,
		Z:           0,
	})
	AddRoom(room)
}

func TestEndToEnd(t *testing.T) {
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

	// Read the login text first.
	loginText, err := reader.ReadString('\r')
	assert.Nil(t, err)
	assert.Equal(t, "Welcome, by what name are you known?"+color.Reset()+"\r", loginText)

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
				if p.Flag("color") {
					assert.Equal(t, color.Parse(response)+"\r", recv)
				} else {
					assert.Equal(t, color.Strip(response)+"\r", recv)
				}

				// Slight cheat here, but easier for testing -- check if prompt
				// should be shown and match for it.
				// Also account for the skip in "Confirm Password" step as the player
				// enters the game.
				if p.ShowPrompt() && test.name != "Confirm Password" {
					recv, err = reader.ReadString('\r')
					assert.Nil(t, err)
					if p.Flag("color") {
						assert.Equal(t, "\n\n"+color.Parse(p.Prompt())+"\r", recv)
					} else {
						assert.Equal(t, "\n\n"+p.Prompt()+"\r", recv)

					}
				}
			}
		})
	}
	os.RemoveAll(t.TempDir())
	server.Close()
}
