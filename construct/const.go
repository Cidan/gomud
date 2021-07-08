package construct

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
