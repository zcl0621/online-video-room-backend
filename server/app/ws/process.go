package ws

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"os"
	"sync"
	"time"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type client struct {
	Conn *websocket.Conn
	Id   int64
}

var fileMap = make(map[string]*os.File)
var fileMapLock = &sync.Mutex{}

var wsMap = make(map[string][]*client)
var wsMapLock = &sync.Mutex{}

func makeVideoFile(key string) (*os.File, error) {
	videoFile, err := os.Create(fmt.Sprintf("/tmp/tmp_%s.txt", key))
	if err != nil {
		log.Println("Failed to create video file:", err)
		return nil, err
	}
	return videoFile, nil
}

func rtcHolder(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println("Failed to upgrade websocket:", err)
		return
	}

	log.Println("New WebSocket connection")

	var request wsRequest
	if err := c.ShouldBindQuery(&request); err != nil {
		log.Println("Failed to read WebSocket message:", err)
		return
	}

	userKey := fmt.Sprintf("%d_%s", request.RoomId, request.Name)
	roomKey := fmt.Sprintf("%d", request.RoomId)
	fileLock := fileMapLock.TryLock()
	if !fileLock {
		sendError(conn, "文件锁失败")
		return
	}
	file, ok := fileMap[userKey]
	if !ok || file == nil {
		nfile, err := makeVideoFile(userKey)
		if err != nil {
			sendError(conn, "创建文件失败")
			return
		}
		file = nfile
		fileMap[userKey] = file
	}
	fileMapLock.Unlock()
	wsc := &client{
		Id:   time.Now().UnixNano(),
		Conn: conn,
	}
	wsLock := wsMapLock.TryLock()
	if !wsLock {
		sendError(conn, "WebSocket锁失败")
		return
	}
	wsList, ok := wsMap[roomKey]
	if !ok || wsList == nil {
		wsList = []*client{
			wsc,
		}
		wsMap[roomKey] = wsList
	} else {
		wsList = append(wsList, wsc)
		wsMap[roomKey] = wsList
	}
	wsMapLock.Unlock()
	closeCH := make(chan struct{}, 100)
	conn.SetCloseHandler(func(code int, text string) error {
		closeCH <- struct{}{}
		return nil
	})
	// Read messages from WebSocket
	for {
		select {
		case <-closeCH:
			log.Println("WebSocket connection closed")
			wsLock := wsMapLock.TryLock()
			if !wsLock {
				sendError(conn, "WebSocket锁失败")
				return
			}
			wsList, ok := wsMap[roomKey]
			if ok && wsList != nil {
				for i, c := range wsList {
					if c.Id == wsc.Id {
						wsList = append(wsList[:i], wsList[i+1:]...)
						wsMap[roomKey] = wsList
						break
					}
				}
			}
			wsMapLock.Unlock()
			return
		default:
			msgType, data, err := conn.ReadMessage()
			if err == nil {
				switch msgType {
				case websocket.TextMessage:
					if string(data) == "end" {
						file.Close()
						go func() {
							convert(fmt.Sprintf("/tmp/tmp_%s.txt", userKey), userKey)
						}()
						return
					} else {
						file.Write([]byte("\n"))
						file.Write(data)
						broadcastVideoData(data, roomKey, wsc.Id)
					}
				}
			}
		}
	}
}

func sendError(conn *websocket.Conn, err string) {
	conn.WriteJSON(&errorResponse{
		Err: err,
	})
}

func broadcastVideoData(data []byte, roomKey string, ownerId int64) {
	for _, client := range wsMap[roomKey] {
		if client != nil && client.Conn != nil && client.Id != ownerId {
			err := client.Conn.WriteMessage(websocket.TextMessage, data)
			if err != nil {
				log.Println("Error broadcasting video data:", err)
				client.Conn.Close()
			}
		}
	}
}
