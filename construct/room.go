package construct

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"os"
	"sync"

	"github.com/Cidan/gomud/config"
	"github.com/Cidan/gomud/path"
	"github.com/rs/zerolog/log"
	uuid "github.com/satori/go.uuid"
)

var exitDirections = []string{"north", "south", "east", "west", "up", "down"}
var inverseDirections = map[string]string{
	"north": "south",
	"south": "north",
	"east":  "west",
	"west":  "east",
	"up":    "down",
	"down":  "up",
}

// RoomExit is an exit to a room. Exits decide state, such as open/closed doors,
// walls, or portals.
type RoomExit struct {
	Direction string
	Name      string
	Door      bool
	Closed    bool
	Locked    bool
	Wall      bool
	Target    string
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
	DirectionExits []*RoomExit
	OtherExits     map[string]*RoomExit
}

// Room is the top level struct for a room.
type Room struct {
	Data *RoomData
	// TODO: Add Direction target pointers to room on load and dig
	players     map[string]*Player
	playerMutex *sync.RWMutex
	exitsMutex  *sync.RWMutex
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
	// The directional exit map should be immutable, never edit
	// the map it self. Concurrent reads of exits is okay.
	exits := make([]*RoomExit, 6)
	for n := range exitDirections {
		exits[n] = new(RoomExit)
	}

	return &Room{
		Data: &RoomData{
			UUID:           uuid.NewV4().String(),
			Name:           "New Room",
			Description:    "This is a new room, with a new description.",
			DirectionExits: exits,
			OtherExits:     make(map[string]*RoomExit),
		},
		players:     make(map[string]*Player),
		playerMutex: new(sync.RWMutex),
		exitsMutex:  new(sync.RWMutex),
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
// This is a "fast" implementation with no allocations. This should be used
// for non-interactive maps, i.e. eagle eye spells, etc.
func (r *Room) Map(radius int64) string {
	str := "\n  "
	startX := r.Data.X - radius
	startY := r.Data.Y + radius
	z := r.Data.Z
	var ry int64 = 0
	for y := startY; y > r.Data.Y-radius; y-- {
		var rx int64 = 0
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
			rx++
		}
		ry++
		str += "\n  "
	}

	return str
}

// WalledMap will generate a map of the area around this room, with walls denoting
// the barriers between rooms.
func (r *Room) WalledMap(radius int64) string {
	gameMap := path.NewMap(radius)
	startX := r.Data.X - radius
	startY := r.Data.Y + radius
	z := r.Data.Z

	var my int64 = 0
	for y := startY; y > r.Data.Y-radius; y-- {
		var mx int64 = 0
		for x := startX; x < r.Data.X+radius; x++ {
			mroom := GetRoom(x, y, z)
			cell := gameMap.Cell(mx, my, z)
			switch {
			case mroom == nil:
				cell.Empty = true
			default:
				mroom.pathAround(cell)
			}
			mx++
		}
		my++
	}
	return gameMap.DrawMap(r.Data.Z)
}

// GeneratePath will generate a path to the target room. Use the path to navigate to the
// given room. This is a heavy implementation and should only be used when a path
// to a room is needed, i.e. hunting another player, mob, or object.
func (r *Room) GeneratePath(target *Room) *path.Path {
	dx := math.Abs(float64(r.Data.X - target.Data.X))
	dy := math.Abs(float64(r.Data.Y - target.Data.Y))
	dz := math.Abs(float64(r.Data.Z - target.Data.Z))

	max := math.Max(dx, dy)
	radius := int64(math.Max(max, dz))

	gameMap := path.NewMap(radius)
	startX := r.Data.X - radius
	startY := r.Data.Y + radius
	startZ := r.Data.Z - radius

	for y := startY; y > r.Data.Y-radius; y-- {
		for x := startX; x < r.Data.X+radius; x++ {
			for z := startZ; z < r.Data.Z+radius; z++ {
				mroom := GetRoom(x, y, z)
				cell := gameMap.Cell(x, y, z)
				switch {
				case mroom == nil:
					cell.Empty = true
				default:
					mroom.pathAround(cell)
				}
			}
		}
	}
	return gameMap.Path(nil, nil)
}

// Exit returns an exit for a given direction.
func (r *Room) Exit(dir string) *RoomExit {
	switch dir {
	case "north":
		return r.Data.DirectionExits[0]
	case "south":
		return r.Data.DirectionExits[1]
	case "east":
		return r.Data.DirectionExits[2]
	case "west":
		return r.Data.DirectionExits[3]
	case "up":
		return r.Data.DirectionExits[4]
	case "down":
		return r.Data.DirectionExits[5]
	default:
		return nil
	}
}

func (r *Room) pathAround(cell *path.Cell) {
	if !r.CanExit("north") {
		cell.Exit("north").Wall = true
	}
	if !r.CanExit("south") {
		cell.Exit("south").Wall = true
	}
	if !r.CanExit("east") {
		cell.Exit("east").Wall = true
	}
	if !r.CanExit("west") {
		cell.Exit("west").Wall = true
	}
	if !r.CanExit("up") {
		cell.Exit("up").Wall = true
	}
	if !r.CanExit("down") {
		cell.Exit("down").Wall = true
	}
}

func (r *Room) IsExit(dir string) bool {
	if r.LinkedRoom(dir) == nil {
		return false
	}
	return true
}

func (r *Room) IsExitWall(dir string) bool {
	if r.LinkedRoom(dir) == nil {
		return true
	}
	return r.Exit(dir).Wall
}

func (r *Room) IsExitClosed(dir string) bool {
	if r.LinkedRoom(dir) == nil {
		return false
	}
	return r.Exit(dir).Closed
}

func (r *Room) IsExitDoor(dir string) bool {
	if r.LinkedRoom(dir) == nil {
		return false
	}
	return r.Exit(dir).Door
}

func (r *Room) CanExit(dir string) bool {
	if r.LinkedRoom(dir) == nil {
		return false
	}
	exit := r.Exit(dir)
	if exit.Closed || exit.Wall {
		return false
	}
	return true
}
