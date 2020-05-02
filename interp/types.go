package interp

type player interface {
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
}

// Interp type for interpreting player commands
type Interp interface {
	Read(string) error
}
