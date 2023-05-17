package server

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"online-video-room-backend/server/app/room"
	"online-video-room-backend/server/app/ws"
	"time"
)

func SetupRouter() *gin.Engine {
	r := gin.New()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	r.Use(cors.New(cors.Config{
		AllowMethods:     []string{"*"},
		AllowHeaders:     []string{"*"},
		ExposeHeaders:    []string{"*"},
		AllowAllOrigins:  true,
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))
	api := r.Group("/api")
	room.InitRoomRouter(api)
	ws.InitWsRouter(api)
	return r
}

func StarServer() {
	r := SetupRouter()
	e := r.Run("0.0.0.0:8080")
	if e != nil {
		panic(e)
	}
}
