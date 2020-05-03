package room

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/Cidan/gomud/atlas"
	"github.com/Cidan/gomud/types"
	uuid "github.com/satori/go.uuid"
)

type Data struct {
	UUID        string
	Name        string
	Description string
	X           int64
	Y           int64
	Z           int64
}

type Room struct {
	Data *Data
}

func New(data *Data) *Room {
	data.UUID = uuid.NewV4().String()

	return &Room{
		Data: data,
	}
}

func (r *Room) GetName() string {
	return r.Data.Name
}

func (r *Room) GetDescription() string {
	return r.Data.Description
}

func (r *Room) GetIndex() string {
	return fmt.Sprintf("%d,%d,%d", r.Data.X, r.Data.Y, r.Data.Z)
}

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
