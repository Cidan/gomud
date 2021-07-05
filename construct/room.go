package construct

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"sync"

	"github.com/Cidan/gomud/config"
	"github.com/rs/zerolog/log"
	uuid "github.com/satori/go.uuid"
)

// RoomData struct for a room. This data is saved to durable storage when a room is
// saved.
type RoomData struct {
	UUID        string
	Name        string
	Description string
	X           int64
	Y           int64
	Z           int64
}

// Room is the top level struct for a room.
type Room struct {
	Data        *RoomData
	Players     map[string]*Player
	playerMutex *sync.RWMutex
}

// LoadRooms loads all the rooms in the world.
func LoadRooms() error {
	os.Mkdir(fmt.Sprintf("%s/rooms", config.GetString("save_path")), 0755)
	files, err := ioutil.ReadDir(fmt.Sprintf("%s/rooms/", config.GetString("save_path")))
	if err != nil {
		return err
	}
	for _, file := range files {
		data, err := ioutil.ReadFile(fmt.Sprintf("%s/rooms/%s", config.GetString("save_path"), file.Name()))
		if err != nil {
			return err
		}
		var roomData RoomData
		err = json.Unmarshal(data, &roomData)
		if err != nil {
			return err
		}
		log.Debug().Str("name", roomData.Name).Msg("loaded room")
		AddRoom(NewRoom(&roomData))
	}
	return nil
}

// NewRoom construct.
func NewRoom(data *RoomData) *Room {
	data.UUID = uuid.NewV4().String()

	return &Room{
		Data:        data,
		Players:     make(map[string]*Player),
		playerMutex: new(sync.RWMutex),
	}
}

// GetName returns the human readable name of a room.
func (r *Room) GetName() string {
	return r.Data.Name
}

// GetDescription returns the human readable description of the room.
func (r *Room) GetDescription() string {
	return r.Data.Description
}

// GetIndex gets the room index as a string.
func (r *Room) GetIndex() string {
	return fmt.Sprintf("%d,%d,%d", r.Data.X, r.Data.Y, r.Data.Z)
}

// Save a room to durable storage.
func (r *Room) Save() error {
	data, err := json.Marshal(r.Data)
	if err != nil {
		return err
	}
	os.Mkdir(fmt.Sprintf("%s/rooms", config.GetString("save_path")), 0755)
	err = ioutil.WriteFile(fmt.Sprintf("%s/rooms/%s", config.GetString("save_path"), r.Data.UUID), data, 0644)
	if err != nil {
		return err
	}
	return nil
}

// LinkedRoom returns a room to which this room can traverse to using
// a direction or portal, given the direction/portal name
func (r *Room) LinkedRoom(dir string) *Room {
	switch dir {
	case "north":
		return GetRoom(r.Data.X, r.Data.Y+1, r.Data.Z)
	case "east":
		return GetRoom(r.Data.X+1, r.Data.Y, r.Data.Z)
	case "south":
		return GetRoom(r.Data.X, r.Data.Y-1, r.Data.Z)
	case "west":
		return GetRoom(r.Data.X-1, r.Data.Y, r.Data.Z)
	case "up":
		return GetRoom(r.Data.X, r.Data.Y, r.Data.Z+1)
	case "down":
		return GetRoom(r.Data.X, r.Data.Y, r.Data.Z-1)
	}
	return nil
}

// AddPlayer adds a player to a room.
func (r *Room) AddPlayer(player *Player) {
	r.playerMutex.Lock()
	defer r.playerMutex.Unlock()
	r.Players[player.GetUUID()] = player
}

// RemovePlayer removes a player from a room.
func (r *Room) RemovePlayer(player *Player) {
	r.playerMutex.Lock()
	defer r.playerMutex.Unlock()
	delete(r.Players, player.GetUUID())
}
