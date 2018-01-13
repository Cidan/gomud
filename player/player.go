package player

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"

	"github.com/Cidan/gomud/types"

	uuid "github.com/satori/go.uuid"
)

// Player construct
type Player struct {
	connection net.Conn
	Input      *bufio.Reader
	Data       *PlayerData
	Interp     types.Interp
	InRoom     *types.Room
}

type PlayerData struct {
	UUID     string
	Name     string
	Password string
}

// New player
func New() *Player {
	pd := &PlayerData{
		UUID: uuid.NewV4().String(),
	}
	return &Player{
		Data: pd,
	}
}

// SetConnection sets the player connection object
func (p *Player) SetConnection(c net.Conn) {
	p.connection = c
	p.Input = bufio.NewReader(c)
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

// Stop a player connection and unload the player from the world.
func (p *Player) Stop() {
	p.connection.Close()
}
