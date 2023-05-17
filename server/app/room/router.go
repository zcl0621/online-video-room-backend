package room

import "github.com/gin-gonic/gin"

func InitRoomRouter(r *gin.RouterGroup) {
	api := r.Group("/room")
	{
		api.POST("/create", createRoom)
		api.POST("/join", joinRoom)
		api.POST("/info", info)
	}
}
