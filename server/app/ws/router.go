package ws

import "github.com/gin-gonic/gin"

func InitWsRouter(g *gin.RouterGroup) {
	api := g.Group("/ws")
	{
		api.GET("/live", rtcHolder)
	}
}
