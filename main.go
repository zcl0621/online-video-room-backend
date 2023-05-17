package main

import (
	"online-video-room-backend/migrate"
	"online-video-room-backend/server"
)

func main() {
	migrate.Migrate()
	server.StarServer()
}
