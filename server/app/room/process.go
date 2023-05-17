package room

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"online-video-room-backend/database"
	"online-video-room-backend/model"
)

func createRoom(c *gin.Context) {
	var request createRequest
	if e := c.ShouldBindJSON(&request); e != nil {
		c.JSON(400, gin.H{"error": "参数错误"})
		return
	}
	roomModel := model.Room{
		MasterName: request.Name,
		GuestName:  "",
	}
	db := database.GetInstance()
	db.Create(&roomModel)
	c.JSON(http.StatusOK, &createResponse{
		RoomId: roomModel.ID,
	})
}

func joinRoom(c *gin.Context) {
	var request joinRequest
	if e := c.ShouldBindJSON(&request); e != nil {
		c.JSON(400, gin.H{"error": "参数错误"})
		return
	}
	db := database.GetInstance()
	var roomModel model.Room
	if e := db.First(&roomModel, request.RoomId).Error; e != nil {
		c.JSON(400, gin.H{"error": "房间不存在"})
		return
	}
	if roomModel.GuestName != "" {
		c.JSON(400, gin.H{"error": "房间已满"})
		return
	}
	roomModel.GuestName = request.Name
	db.Save(&roomModel)
	c.JSON(http.StatusOK, &joinResponse{
		RoomId: roomModel.ID,
	})
}

func info(c *gin.Context) {
	var request infoRequest
	if e := c.ShouldBindJSON(&request); e != nil {
		c.JSON(400, gin.H{"error": "参数错误"})
		return
	}
	db := database.GetInstance()
	var roomModel model.Room
	if e := db.First(&roomModel, request.RoomId).Error; e != nil {
		c.JSON(400, gin.H{"error": "房间不存在"})
		return
	}
	c.JSON(http.StatusOK, &infoResponse{
		MasterName: roomModel.MasterName,
		GuestName:  roomModel.GuestName,
	})
}
