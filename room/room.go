package room

type RoomData struct {
	Name        string
	Description string
	X           int64
	Y           int64
	Z           int64
}

type Room struct {
	Data *RoomData
}

func New(data *RoomData) *Room {
	return &Room{
		Data: data,
	}
}

func (r *Room) GetName() string {
	return r.Data.Name
}

func (r *Room) GetDescription() string {
	return r.Data.Description
}
