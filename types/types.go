package types

type Interp interface {
	Read(string) error
}

type Room interface {
	GetName() string
}

type Player interface {
	Write(string, ...string) error
	Save() error
}
