package types

// The types package makes it easier to reason about circular imports.
// Basic constructs, such as players, objects, and rooms all have complex
// and multiple interactions with one another.
//
// This package must never import an external dependency from this project.
// Outside modules (uuid, net.Conn, etc) are okay, as they will never reference
// this package.

// Player defines the interface for a player object.
type Player interface {
	Load() (bool, error)
	Write(string, ...interface{})
	SetName(string)
	Stop()
	IsPassword(string) bool
	SetPassword(string)
	GetName() string
	SetInterp(Interp)
	GetUUID() string
	Save() error
	ToRoom(Room) bool
	Command(string) error
	GetRoom() Room
	Buffer(string, ...interface{})
	Flush()
}

// Interp type for interpreting player commands.
type Interp interface {
	Read(string) error
}

// Room defines the interface for a room object.
type Room interface {
	GetName() string
	GetDescription() string
}
