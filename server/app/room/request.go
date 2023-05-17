package room

type createRequest struct {
	Name string `json:"name" binding:"required"`
}

type joinRequest struct {
	Name   string `json:"name" binding:"required"`
	RoomId uint   `json:"room_id" binding:"required"`
}

type infoRequest struct {
	RoomId uint `json:"room_id" binding:"required"`
}
