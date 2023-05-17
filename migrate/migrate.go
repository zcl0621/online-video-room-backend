package migrate

import (
	"online-video-room-backend/database"
	"online-video-room-backend/model"
)

func Migrate() {
	db := database.GetInstance()
	db.AutoMigrate(&model.Room{})
}
