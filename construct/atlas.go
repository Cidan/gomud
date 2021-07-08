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

type direction int

const (
	dirNorth direction = iota
	dirSouth
	dirEast
	dirWest
	dirUp
	dirDown
)

var exitDirections = []direction{dirNorth, dirSouth, dirEast, dirWest, dirUp, dirDown}
var inverseDirections = map[direction]direction{
	dirNorth: dirSouth,
	dirSouth: dirNorth,
	dirEast:  dirWest,
	dirWest:  dirEast,
	dirUp:    dirDown,
	dirDown:  dirUp,
}

var dirNames = map[direction]string{
	dirNorth: "north",
	dirSouth: "south",
	dirEast:  "east",
	dirWest:  "west",
	dirUp:    "up",
	dirDown:  "down",
}

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

func getRelativeDir(dir direction) (x, y, z int64) {
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

func dirToName(dir direction) string {
	return dirNames[dir]
}

func MakeDefaultRoomSet() {
	room := NewRoom()
	room.SetName("The Alpha")
	room.SetDescription("It all starts here.")
	room.SetCoordinates(0, 0, 0)
	for _, dir := range exitDirections {
		room.Exit(dir).Wall = true
	}

	err := room.Save()
	if err != nil {
		panic(err)
	}
	AddRoom(room)
}
