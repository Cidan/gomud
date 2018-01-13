package atlas

import (
	"fmt"
	"sync"

	"github.com/Cidan/gomud/world"
)

var worldMap map[string]*world.Room
var worldLock sync.RWMutex

func SetupWorld() {
	worldMap = make(map[string]*world.Room)
}

func genRoomIndex(X, Y, Z int64) string {
	return fmt.Sprintf("%d,%d,%d", X, Y, Z)
}

func getRoomIndex(r *world.Room) string {
	return genRoomIndex(r.Data.X, r.Data.Y, r.Data.Z)
}

func GetRoom(X, Y, Z int64) *world.Room {
	defer worldLock.RUnlock()
	worldLock.RLock()
	return worldMap[genRoomIndex(X, Y, Z)]
}

func AddRoom(r *world.Room) {
	defer worldLock.Unlock()
	worldLock.Lock()
	worldMap[getRoomIndex(r)] = r
}
