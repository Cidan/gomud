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

// RoomExit is an exit to a room. Exits decide state, such as open/closed doors,
// walls, or portals.
type RoomExit struct {
	Direction string
	Closed    bool
	Locked    bool
	Wall      bool
}

// RoomData struct for a room. This data is saved to durable storage when a room is
// saved.
type RoomData struct {
	UUID           string
	Name           string
	Description    string
	X              int64
	Y              int64
	Z              int64
	DirectionExits map[string]*RoomExit
	OtherExits     map[string]*RoomExit
}

// Room is the top level struct for a room.
type Room struct {
	Data        *RoomData
	players     map[string]*Player
	playerMutex *sync.RWMutex
}

// PlayerList is the callback function signature for listing players in a room.
type PlayerList func(string, *Player)

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
		room := NewRoom()
		err = json.Unmarshal(data, &room.Data)
		if err != nil {
			return err
		}
		log.Debug().Str("name", room.GetName()).Msg("loaded room")
		AddRoom(room)
		// Save the room after load in order to apply any possible migrations.
		if err := room.Save(); err != nil {
			return err
		}
	}
	return nil
}

// NewRoom construct.
func NewRoom() *Room {
	return &Room{
		Data: &RoomData{
			UUID:           uuid.NewV4().String(),
			Name:           "New Room",
			Description:    "This is a new room, with a new description.",
			DirectionExits: make(map[string]*RoomExit),
			OtherExits:     make(map[string]*RoomExit),
		},
		players:     make(map[string]*Player),
		playerMutex: new(sync.RWMutex),
	}
}

// GetName returns the human readable name of a room.
func (r *Room) GetName() string {
	return r.Data.Name
}

//SetName sets the name of this room.
func (r *Room) SetName(name string) {
	r.Data.Name = name
}

// GetDescription returns the human readable description of the room.
func (r *Room) GetDescription() string {
	return r.Data.Description
}

// SetDescription sets the description of this room.
func (r *Room) SetDescription(desc string) {
	r.Data.Description = desc
}

func (r *Room) SetCoordinates(x, y, z int64) {
	r.Data.X = x
	r.Data.Y = y
	r.Data.Z = z
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
	r.players[player.GetUUID()] = player
}

// RemovePlayer removes a player from a room.
func (r *Room) RemovePlayer(player *Player) {
	r.playerMutex.Lock()
	defer r.playerMutex.Unlock()
	delete(r.players, player.GetUUID())
}

// AllPlayers loops through all players in a room and runs the callback function
// for each player in a concurrent safe manner.
func (r *Room) AllPlayers(fn PlayerList) {
	r.playerMutex.RLock()
	// Compile a list of players locally and use that as the interator.
	// This is done so that long running callback functions are localized
	// and don't block room entrances.
	var plist []*Player
	for _, p := range r.players {
		plist = append(plist, p)
	}
	r.playerMutex.RUnlock()

	for _, p := range plist {
		fn(p.GetUUID(), p)
	}
}

// Map generates a map with this room at the center, with the given radius.
func (r *Room) Map(radius int64) string {
	str := "\n  "
	startX := r.Data.X - radius
	startY := r.Data.Y + radius
	z := r.Data.Z
	for y := startY; y > r.Data.Y-radius; y-- {
		for x := startX; x < r.Data.X+radius; x++ {
			mroom := GetRoom(x, y, z)
			switch {
			case mroom == nil:
				str += " "
			case mroom == r:
				str += "{R*{x"
			default:
				str += "{W#{x"
			}
		}
		str += "\n  "
	}
	return str
}
