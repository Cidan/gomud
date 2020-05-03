package atlas

import (
	"fmt"
	"io/ioutil"

	"github.com/Cidan/gomud/types"
)

var worldMap map[string]types.Room

func LoadWorld() error {
	if worldMap == nil {
		worldMap = make(map[string]types.Room)
	}

	files, err := ioutil.ReadDir("/tmp/rooms/")
	if err != nil {
		return err
	}

	for _, file := range files {
		file.Name()
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
}

func IsRoom(X, Y, Z int64) types.Room {
	room, _ := worldMap[genRoomIndex(X, Y, Z)]
	return room
}
