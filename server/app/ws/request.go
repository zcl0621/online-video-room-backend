package ws

type wsRequest struct {
	RoomId uint   `json:"room_id" form:"room_id" binding:"required"`
	Name   string `json:"name" form:"name" binding:"required"`
}
