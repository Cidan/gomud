package construct

import (
	"context"
	"fmt"
	"sync"
)

type AtlasData struct {
	worldMap          map[string]*Room
	worldRoomUUID     map[string]*Room
	allPlayers        map[string]*Player
	worldSize         int64
	worldMapMutex     sync.RWMutex
	worldRoomMutex    sync.RWMutex
	allPlayersMutex   sync.RWMutex
	roomModifierMutex sync.Mutex
}

var Atlas *AtlasData

func init() {
	Atlas = &AtlasData{
		worldMap:          make(map[string]*Room),
		worldRoomUUID:     make(map[string]*Room),
		allPlayers:        make(map[string]*Player),
		worldMapMutex:     sync.RWMutex{},
		worldRoomMutex:    sync.RWMutex{},
		allPlayersMutex:   sync.RWMutex{},
		roomModifierMutex: sync.Mutex{},
	}
}

func (a *AtlasData) genRoomIndex(X, Y, Z int64) string {
	return fmt.Sprintf("%d,%d,%d", X, Y, Z)
}

// GetRoom returns a pointer to a room at the given coordinates.
func (a *AtlasData) GetRoom(X, Y, Z int64) *Room {
	a.worldMapMutex.RLock()
	defer a.worldMapMutex.RUnlock()
	if room, ok := a.worldMap[a.genRoomIndex(X, Y, Z)]; ok {
		return room
	}
	return nil
}

// GetRoomByUUID returns a pointer to a room via the room UUID.
func (a *AtlasData) GetRoomByUUID(uuid string) *Room {
	a.worldRoomMutex.RLock()
	defer a.worldRoomMutex.RUnlock()
	if room, ok := a.worldRoomUUID[uuid]; ok {
		return room
	}
	return nil
}

// AddRoom instatiates a room into the game world.
func (a *AtlasData) AddRoom(r *Room) {
	a.worldMapMutex.Lock()
	a.worldMap[r.GetIndex()] = r
	a.worldMapMutex.Unlock()

	a.worldRoomMutex.Lock()
	a.worldRoomUUID[r.Data.UUID] = r
	a.worldSize++
	a.worldRoomMutex.Unlock()
}

// AddPlayer adds a player to the global game state. Returns existing
// player reference if the player already exists globally.
func (a *AtlasData) AddPlayer(ctx context.Context, p *Player) *Player {
	a.allPlayersMutex.Lock()
	defer a.allPlayersMutex.Unlock()
	if existingPlayer, ok := a.allPlayers[p.GetName(ctx)]; ok {
		return existingPlayer
	}
	a.allPlayers[p.GetName(ctx)] = p
	return nil
}

func (a *AtlasData) RemovePlayer(ctx context.Context, p *Player) {
	a.allPlayersMutex.Lock()
	defer a.allPlayersMutex.Unlock()
	delete(a.allPlayers, p.GetName(ctx))
}

// WorldSize returns the number of rooms in the world.
func (a *AtlasData) WorldSize() int64 {
	return a.worldSize
}

func (a *AtlasData) getRelativeDir(dir direction) (x, y, z int64) {
	switch dir {
	case dirNorth:
		return 0, 1, 0
	case dirSouth:
		return 0, -1, 0
	case dirEast:
		return 1, 0, 0
	case dirWest:
		return -1, 0, 0
	case dirUp:
		return 0, 0, 1
	case dirDown:
		return 0, 0, -1
	}
	return 0, 0, 0
}

func (a *AtlasData) dirToName(dir direction) string {
	return dirNames[dir]
}

func (a *AtlasData) MakeDefaultRoomSet(ctx context.Context) {
	room := NewRoom()
	room.SetName("The Alpha")
	room.SetDescription("It all starts here.")
	room.SetCoordinates(0, 0, 0)
	for _, dir := range exitDirections {
		room.Exit(ctx, dir).Wall = true
	}

	err := room.Save()
	if err != nil {
		panic(err)
	}
	a.AddRoom(room)
}
