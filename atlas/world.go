package atlas

import (
	"fmt"

	"github.com/Cidan/gomud/room"
	cmap "github.com/orcaman/concurrent-map"
)

var worldMap cmap.ConcurrentMap

func SetupWorld() {
	worldMap = cmap.New()
}

func genRoomIndex(X, Y, Z int64) string {
	return fmt.Sprintf("%d,%d,%d", X, Y, Z)
}

func getRoomIndex(r *room.Room) string {
	return genRoomIndex(r.Data.X, r.Data.Y, r.Data.Z)
}

func GetRoom(X, Y, Z int64) *room.Room {
	if tmp, ok := worldMap.Get(genRoomIndex(X, Y, Z)); ok {
		return tmp.(*room.Room)
	}
	return nil
}

func AddRoom(r *room.Room) {
	worldMap.Set(getRoomIndex(r), r)
}
