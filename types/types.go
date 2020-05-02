package types

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
}

// Interp type for interpreting player commands
type Interp interface {
	Read(string) error
}

// Room defines the interface for a room object.
type Room interface {
	GetName() string
}
