package construct

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"os"
	"strings"

	"github.com/Cidan/gomud/config"
	"github.com/Cidan/gomud/lock"
	"github.com/Cidan/gomud/path"
	"github.com/rs/zerolog/log"
	uuid "github.com/satori/go.uuid"
)

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
	Data      *RoomData
	exitRooms []*Room
	players   map[string]*Player
	lock      *lock.Lock
}

// PlayerList is the callback function signature for listing players in a room.
type PlayerList func(string, *Player)

// LoadRooms loads all the rooms in the world.
func LoadRooms(ctx context.Context) error {
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
		Atlas.AddRoom(room)
		// Save the room after load in order to apply any possible migrations.
		if err := room.Save(); err != nil {
			return err
		}
	}

	// Loop through all rooms and link their exits in memory. Somewhat expensive
	// in large worlds, but only needs to be done once. This allows for fast room
	// movement without global lookups.
	for _, room := range Atlas.worldMap {
		for _, dir := range exitDirections {
			exit := room.Exit(ctx, dir)
			if exit.Target != "" {
				room.exitRooms[dir] = Atlas.GetRoomByUUID(exit.Target)
			}
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

	uuid := uuid.NewV4().String()
	return &Room{
		Data: &RoomData{
			UUID:           uuid,
			Name:           "New Room",
			Description:    "This is a new room, with a new description.",
			DirectionExits: exits,
			OtherExits:     make(map[string]*RoomExit),
		},
		exitRooms: make([]*Room, 6),
		players:   make(map[string]*Player),
		lock:      lock.New(uuid),
	}
}

// Delete will delete a room from the world, unlinking it from
// the game, and shunting players in a random open direction.
// If no direction is available, the player is returned to 0,0,0
// for now, until home rooms are implemented.
func (r *Room) Delete(ctx context.Context) error {
	r.lock.Lock(ctx)
	defer r.lock.Unlock(ctx)
	// Lock globally when deleting a room. This prevents a race where
	// multiple rooms may be deleted at once, causing weird races where
	// players would not exist in a room at all.
	Atlas.roomModifierMutex.Lock()
	defer Atlas.roomModifierMutex.Unlock()
	var toRoom *Room

	for dir, exitRoom := range r.exitRooms {
		if exitRoom == nil {
			continue
		}
		// Isolate entry from the target room from other directions.
		inverse := inverseDirections[exitDirections[dir]]
		exitRoom.Exit(ctx, inverse).Target = ""
		exitRoom.Exit(ctx, inverse).Closed = false
		exitRoom.Exit(ctx, inverse).Wall = false
		exitRoom.Exit(ctx, inverse).Locked = false
		exitRoom.Exit(ctx, inverse).Name = ""
		toRoom = exitRoom
		exitRoom.exitRooms[dir] = nil

		// Isolate this room from other entries.
		exit := r.Exit(ctx, exitDirections[dir])
		exit.Closed = false
		exit.Door = false
		exit.Wall = false
		exit.Target = ""
		r.exitRooms[dir] = nil
	}

	// Move all player to an adjecent room.
	if toRoom != nil {
		for _, p := range r.players {
			p.ToRoom(ctx, toRoom)
		}
	}

	return nil
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
func (r *Room) LinkedRoom(ctx context.Context, dir direction) *Room {
	r.lock.Lock(ctx)
	defer r.lock.Unlock(ctx)
	return r.exitRooms[dir]
}

// PhysicalRoom returns a room which exists physically in the direction
// given based on the coordinate map.
func (r *Room) PhysicalRoom(dir direction) *Room {
	switch dir {
	case dirNorth:
		return Atlas.GetRoom(r.Data.X, r.Data.Y+1, r.Data.Z)
	case dirEast:
		return Atlas.GetRoom(r.Data.X+1, r.Data.Y, r.Data.Z)
	case dirSouth:
		return Atlas.GetRoom(r.Data.X, r.Data.Y-1, r.Data.Z)
	case dirWest:
		return Atlas.GetRoom(r.Data.X-1, r.Data.Y, r.Data.Z)
	case dirUp:
		return Atlas.GetRoom(r.Data.X, r.Data.Y, r.Data.Z+1)
	case dirDown:
		return Atlas.GetRoom(r.Data.X, r.Data.Y, r.Data.Z-1)
	}
	return nil
}

// AddPlayer adds a player to a room.
func (r *Room) AddPlayer(ctx context.Context, player *Player) {
	r.lock.Lock(ctx)
	defer r.lock.Unlock(ctx)
	r.players[player.GetUUID(ctx)] = player
}

// RemovePlayer removes a player from a room.
func (r *Room) RemovePlayer(ctx context.Context, player *Player) {
	r.lock.Lock(ctx)
	defer r.lock.Unlock(ctx)
	delete(r.players, player.GetUUID(ctx))
}

// AllPlayers loops through all players in a room and runs the callback function
// for each player in a concurrent safe manner.
func (r *Room) AllPlayers(ctx context.Context, fn PlayerList) {
	r.lock.Lock(ctx)
	// Compile a list of players locally and use that as the interator.
	// This is done so that long running callback functions are localized
	// and don't block room entrances.
	var plist []*Player
	for _, p := range r.players {
		plist = append(plist, p)
	}
	r.lock.Unlock(ctx)

	for _, p := range plist {
		fn(p.GetUUID(ctx), p)
	}
}

func (r *Room) GetPlayer(ctx context.Context, prefix string) *Player {
	r.lock.Lock(ctx)
	defer r.lock.Unlock(ctx)
	for _, p := range r.players {
		if strings.HasPrefix(p.GetName(ctx), prefix) {
			return p
		}
	}
	return nil
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
			mroom := Atlas.GetRoom(x, y, z)
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
func (r *Room) WalledMap(ctx context.Context, radius int64) string {
	gameMap := path.NewMap(radius)
	startX := r.Data.X - radius
	startY := r.Data.Y + radius
	z := r.Data.Z

	var my int64 = 0
	for y := startY; y > r.Data.Y-radius; y-- {
		var mx int64 = 0
		for x := startX; x < r.Data.X+radius; x++ {
			mroom := Atlas.GetRoom(x, y, z)
			cell := gameMap.Cell(mx, my, z)
			switch {
			case mroom == nil:
				cell.Empty = true
			default:
				mroom.pathAround(ctx, cell)
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
func (r *Room) GeneratePath(ctx context.Context, target *Room) *path.Path {
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
				mroom := Atlas.GetRoom(x, y, z)
				cell := gameMap.Cell(x, y, z)
				switch {
				case mroom == nil:
					cell.Empty = true
				default:
					mroom.pathAround(ctx, cell)
				}
			}
		}
	}
	return gameMap.Path(nil, nil)
}

// Set an exit room for this direction.
func (r *Room) SetExitRoom(ctx context.Context, dir direction, target *Room) {
	r.lock.Lock(ctx)
	defer r.lock.Unlock(ctx)
	r.Exit(ctx, dir).Target = target.Data.UUID
	r.exitRooms[dir] = target
}

// Exit returns an exit for a given direction.
func (r *Room) Exit(ctx context.Context, dir direction) *RoomExit {
	r.lock.Lock(ctx)
	defer r.lock.Unlock(ctx)
	return r.Data.DirectionExits[dir]
}

func (r *Room) pathAround(ctx context.Context, cell *path.Cell) {
	for _, dir := range exitDirections {
		if !r.CanExit(ctx, dir) {
			cell.Exit(Atlas.dirToName(dir)).Wall = true
		}
	}
}

func (r *Room) IsExit(ctx context.Context, dir direction) bool {
	r.lock.Lock(ctx)
	defer r.lock.Unlock(ctx)
	if r.LinkedRoom(ctx, dir) == nil {
		return false
	}
	return true
}

func (r *Room) IsExitWall(ctx context.Context, dir direction) bool {
	r.lock.Lock(ctx)
	defer r.lock.Unlock(ctx)

	if r.LinkedRoom(ctx, dir) == nil {
		return true
	}
	return r.Exit(ctx, dir).Wall
}

func (r *Room) IsExitClosed(ctx context.Context, dir direction) bool {
	r.lock.Lock(ctx)
	defer r.lock.Unlock(ctx)

	if r.LinkedRoom(ctx, dir) == nil {
		return false
	}
	return r.Exit(ctx, dir).Closed
}

func (r *Room) IsExitDoor(ctx context.Context, dir direction) bool {
	r.lock.Lock(ctx)
	defer r.lock.Unlock(ctx)

	if r.LinkedRoom(ctx, dir) == nil {
		return false
	}
	return r.Exit(ctx, dir).Door
}

func (r *Room) CanExit(ctx context.Context, dir direction) bool {
	r.lock.Lock(ctx)
	defer r.lock.Unlock(ctx)

	if r.LinkedRoom(ctx, dir) == nil {
		return false
	}
	exit := r.Exit(ctx, dir)
	if exit.Closed || exit.Wall {
		return false
	}
	return true
}
