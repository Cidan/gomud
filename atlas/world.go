package atlas

import (
	"fmt"

	"github.com/Cidan/gomud/types"
)

var worldMap map[string]types.Room
var worldSize int64

func SetupWorld() error {
	if worldMap == nil {
		worldMap = make(map[string]types.Room)
	}
	return nil
}

func genRoomIndex(X, Y, Z int64) string {
	return fmt.Sprintf("%d,%d,%d", X, Y, Z)
}

func GetRoom(X, Y, Z int64) types.Room {
	if room, ok := worldMap[genRoomIndex(X, Y, Z)]; ok {
		return room
	}
	return nil
}

func AddRoom(r types.Room) {
	worldMap[r.GetIndex()] = r
	worldSize++
}

func IsRoom(X, Y, Z int64) types.Room {
	room, _ := worldMap[genRoomIndex(X, Y, Z)]
	return room
}

func WorldSize() int64 {
	return worldSize
}
