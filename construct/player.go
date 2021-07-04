package construct

import (
	"bufio"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"strings"

	"github.com/rs/zerolog/log"

	uuid "github.com/satori/go.uuid"
)

func hashPassword(pw string) string {
	h := sha512.New()
	io.WriteString(h, pw)
	return hex.EncodeToString(h.Sum(nil))
}

// Player construct
type Player struct {
	connection    net.Conn
	input         *bufio.Reader
	Data          *playerData
	gameInterp    *Game
	buildInterp   *BuildInterp
	loginInterp   *Login
	currentInterp Interp
	inRoom        *Room
	textBuffer    string
}

// This is the main data construct for a human player. Any new flags, attributes
// or other data that needs to carry over between sessions, goes here. Do not
// use this field as storage for temporary variables. Use the Player struct
// above for temporary data that does not need to be saved.
// Additionally, all player fields must be exported in order to be saved.
type playerData struct {
	UUID     string
	Name     string
	Password string
}

// NewPlayer constructs a new player
func NewPlayer() *Player {

	p := &Player{
		Data: &playerData{
			UUID: uuid.NewV4().String(),
		},
	}

	return p
}

// SetConnection sets the player connection object
func (p *Player) SetConnection(c net.Conn) {
	p.connection = c
	p.input = bufio.NewReader(c)
}

// Start this player and their interp loop.
func (p *Player) Start() {
	p.buildInterp = NewBuildInterp(p)
	p.gameInterp = NewGameInterp(p)
	p.loginInterp = NewLoginInterp(p)
	// TODO: Eventually split this line out to another function.
	p.SetInterp(p.loginInterp)

	p.Write("Welcome, by what name are you known?")

	for {
		str, err := p.input.ReadString('\n')
		if err != nil {
			log.Error().Err(err).Msg("Error reading player input.")
			p.Stop()
			break
		}
		str = strings.TrimSpace(str)
		err = p.currentInterp.Read(str)
		switch err {
		case ErrCommandNotFound:
			p.Write("Huh?")
		case nil:
			break
		default:
			log.Error().Err(err).
				Str("player", p.Data.UUID).
				Msg("Error reading input from player.")
			log.Debug().Msg(str)
		}
	}
}

// SetInterp for a player.
func (p *Player) SetInterp(i Interp) {
	p.currentInterp = i
}

// Buffer will buffer output text until Flush() is called.
func (p *Player) Buffer(text string, args ...interface{}) {
	p.textBuffer += fmt.Sprintf(text, args...)
}

// Flush will write the player buffer to the player and clear the buffer.
func (p *Player) Flush() {
	fmt.Fprint(p.connection, p.textBuffer)
	p.textBuffer = ""
}

// Write output to a player.
func (p *Player) Write(text string, args ...interface{}) {
	fmt.Fprintf(p.connection, text, args...)
}

// Save a player to disk
// TODO: just /tmp for now
func (p *Player) Save() error {
	data, err := json.Marshal(p.Data)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile("/tmp/"+p.Data.Name, data, 0644)
	if err != nil {
		return err
	}
	return nil
}

// Load a player from source. Returns true if player was loaded.
func (p *Player) Load() (bool, error) {
	// TODO: This is absurdly unsafe. Fix this.
	data, err := ioutil.ReadFile("/tmp/" + p.Data.Name)
	// TODO: Make this more robust, need to know if error is because of file
	// not found, or error reading.
	if err != nil {
		log.Error().Err(err).Str("player", p.Data.Name).Msg("error loading player")
		return false, nil
	}

	err = json.Unmarshal(data, &p.Data)
	if err != nil {
		return false, err
	}

	return true, nil
}

// Stop a player connection and unload the player from the world.
func (p *Player) Stop() {
	p.connection.Close()
}

// ToRoom moves a player to a room
// TODO: Eventually, unwind combat, etc.
func (p *Player) ToRoom(target *Room) bool {
	p.inRoom = target
	return true
}

// Command runs a command through the interp for the player.
func (p *Player) Command(cmd string) error {
	return p.currentInterp.Read(cmd)
}

// GetUUID of a player.
func (p *Player) GetUUID() string {
	return p.Data.UUID
}

// GetName of a player.
func (p *Player) GetName() string {
	return p.Data.Name
}

// SetName to a player.
func (p *Player) SetName(name string) {
	p.Data.Name = name
}

// IsPassword takes an unhashed string and returns true if the input matches
// the user password.
func (p *Player) IsPassword(password string) bool {
	return p.Data.Password == hashPassword(password)
}

// SetPassword takes an unhashed string and sets that as the user password.
func (p *Player) SetPassword(password string) {
	p.Data.Password = hashPassword(password)
	return
}

// GetRoom returns the room the player is currently in
func (p *Player) GetRoom() *Room {
	return p.inRoom
}
