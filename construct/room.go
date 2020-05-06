package construct

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/Cidan/gomud/atlas"
	"github.com/Cidan/gomud/types"
	"github.com/rs/zerolog/log"
	uuid "github.com/satori/go.uuid"
)

// Data struct for a room. This data is saved to durable storage when a room is
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
	Data *RoomData
}

// LoadAll loads all the rooms in the world.
func LoadRooms() error {
	files, err := ioutil.ReadDir("/tmp/rooms/")
	if err != nil {
		return err
	}
	for _, file := range files {
		data, err := ioutil.ReadFile("/tmp/rooms/" + file.Name())
		if err != nil {
			return err
		}
		var roomData RoomData
		err = json.Unmarshal(data, &roomData)
		if err != nil {
			return err
		}
		log.Debug().Str("name", roomData.Name).Msg("loaded room")
		atlas.AddRoom(NewRoom(&roomData))
	}
	return nil
}

// New room construct.
func NewRoom(data *RoomData) *Room {
	data.UUID = uuid.NewV4().String()

	return &Room{
		Data: data,
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
	err = ioutil.WriteFile("/tmp/rooms/"+r.Data.UUID, data, 0644)
	if err != nil {
		return err
	}
	return nil
}

// LinkedRoom returns a room to which this room can traverse to using
// a direction or portal, given the direction/portal name
func (r *Room) LinkedRoom(dir string) types.Room {
	switch dir {
	case "north":
		return atlas.IsRoom(r.Data.X, r.Data.Y+1, r.Data.Z)
	case "east":
		return atlas.IsRoom(r.Data.X+1, r.Data.Y, r.Data.Z)
	case "south":
		return atlas.IsRoom(r.Data.X, r.Data.Y-1, r.Data.Z)
	case "west":
		return atlas.IsRoom(r.Data.X-1, r.Data.Y, r.Data.Z)
	case "up":
		return atlas.IsRoom(r.Data.X, r.Data.Y, r.Data.Z+1)
	case "down":
		return atlas.IsRoom(r.Data.X, r.Data.Y, r.Data.Z-1)
	}
	return nil
}
