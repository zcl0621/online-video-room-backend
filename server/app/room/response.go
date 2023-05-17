package room

type createResponse struct {
	RoomId uint `json:"room_id"`
}

type joinResponse struct {
	RoomId uint `json:"room_id"`
}

type infoResponse struct {
	MasterName string `json:"master_name"`
	GuestName  string `json:"guest_name"`
}
