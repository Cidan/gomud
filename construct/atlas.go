package construct

import (
	"fmt"
)

var worldMap map[string]*Room
var worldSize int64

func SetupWorld() error {
	if worldMap == nil {
		worldMap = make(map[string]*Room)
	}
	return nil
}

func genRoomIndex(X, Y, Z int64) string {
	return fmt.Sprintf("%d,%d,%d", X, Y, Z)
}

func GetRoom(X, Y, Z int64) *Room {
	if room, ok := worldMap[genRoomIndex(X, Y, Z)]; ok {
		return room
	}
	return nil
}

func AddRoom(r *Room) {
	worldMap[r.GetIndex()] = r
	worldSize++
}

func IsRoom(X, Y, Z int64) *Room {
	room, _ := worldMap[genRoomIndex(X, Y, Z)]
	return room
}

func WorldSize() int64 {
	return worldSize
}
