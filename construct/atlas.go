package construct

import (
	"fmt"
	"sync"
)

var worldMap map[string]*Room
var worldRoomUUID map[string]*Room
var worldSize int64
var worldMapMutex sync.RWMutex
var worldRoomMutex sync.RWMutex

func init() {
	worldMap = make(map[string]*Room)
	worldRoomUUID = make(map[string]*Room)
	worldMapMutex = sync.RWMutex{}
	worldRoomMutex = sync.RWMutex{}
}

func genRoomIndex(X, Y, Z int64) string {
	return fmt.Sprintf("%d,%d,%d", X, Y, Z)
}

// GetRoom returns a pointer to a room at the given coordinates.
func GetRoom(X, Y, Z int64) *Room {
	worldMapMutex.RLock()
	defer worldMapMutex.RUnlock()
	if room, ok := worldMap[genRoomIndex(X, Y, Z)]; ok {
		return room
	}
	return nil
}

// GetRoomByUUID returns a pointer to a room via the room UUID.
func GetRoomByUUID(uuid string) *Room {
	worldRoomMutex.RLock()
	defer worldRoomMutex.RUnlock()
	if room, ok := worldRoomUUID[uuid]; ok {
		return room
	}
	return nil
}

// AddRoom instatiates a room into the game world.
func AddRoom(r *Room) {
	worldMapMutex.Lock()
	worldMap[r.GetIndex()] = r
	worldMapMutex.Unlock()

	worldRoomMutex.Lock()
	worldRoomUUID[r.Data.UUID] = r
	worldSize++
	worldRoomMutex.Unlock()
}

// WorldSize returns the number of rooms in the world.
func WorldSize() int64 {
	return worldSize
}

func getRelativeDir(dir string) (x, y, z int64) {
	switch dir {
	case "north":
		return 0, 1, 0
	case "south":
		return 0, -1, 0
	case "east":
		return 1, 0, 0
	case "west":
		return -1, 0, 0
	case "up":
		return 0, 0, 1
	case "down":
		return 0, 0, -1
	}
	return 0, 0, 0
}
