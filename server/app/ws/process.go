package ws

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/pion/webrtc/v3"
	"log"
	"net/http"
	"online-video-room-backend/database"
	"online-video-room-backend/model"
	"os"
	"os/signal"
	"sync"
)

func init() {
	go closePeerConnections()
}

var peerConnections = make(map[string]*webrtc.PeerConnection)
var perrLock = &sync.Mutex{}

func rtcHolder(c *gin.Context) {
	var request wsRequest
	if e := c.ShouldBindQuery(&request); e != nil {
		c.JSON(400, gin.H{"error": "参数错误"})
		return
	}
	conn, err := (&websocket.Upgrader{
		ReadBufferSize:  10240,
		WriteBufferSize: 10240,
		// 允许所有CORS跨域请求
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}).Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		c.JSON(400, gin.H{"error": "连接异常"})
		return
	}
	db := database.GetInstance()
	var roomModel model.Room
	if e := db.First(&roomModel, request.RoomId).Error; e != nil {
		sendMsg(conn, []byte("房间不存在"))
		return
	}
	if roomModel.GuestName != request.Name || roomModel.MasterName != request.Name {
		sendMsg(conn, []byte("不是房间用户"))
		return
	}
	ok := perrLock.TryLock()
	if !ok {
		sendMsg(conn, []byte("连接数已满"))
		return
	}
	key := fmt.Sprintf("%d_%s", request.RoomId, request.Name)
	pc, ok := peerConnections[key]
	if !ok {
		peerConnection, err := webrtc.NewPeerConnection(webrtc.Configuration{})
		if err != nil {
			sendMsg(conn, []byte("创建webrtc失败"))
			return
		}
		peerConnections[key] = peerConnection
		pc = peerConnection
	}
	perrLock.Unlock()
	if pc == nil {
		sendMsg(conn, []byte("创建webrtc失败"))
		return
	}
	// Set up ICE handling
	pc.OnICECandidate(func(candidate *webrtc.ICECandidate) {
		if candidate != nil {
			// Send ICE candidate to remote peer through WebSocket
			if err := conn.WriteJSON(candidate.ToJSON()); err != nil {
				log.Println("Failed to send ICE candidate:", err)
			}
		}
	})

	// Set up SDP handling
	pc.OnTrack(func(track *webrtc.TrackRemote, receiver *webrtc.RTPReceiver) {
		log.Println("New remote track:", track.ID())

		// Create a file to record the video
		file, err := os.Create(fmt.Sprintf("%s.webm", key))
		if err != nil {
			log.Println("Failed to create video file:", err)
			return
		}

		// Write the received video packets to the file
		rtpRecorder := &RTPRecorder{
			Track: track,
			File:  file,
		}
		rtpRecorder.Start()
	})

	// Handle SDP offer/answer
	handleSDP := func(sessionDescription webrtc.SessionDescription) {
		if err := pc.SetRemoteDescription(sessionDescription); err != nil {
			log.Println("Failed to set remote description:", err)
			return
		}

		if sessionDescription.Type == webrtc.SDPTypeOffer {
			// Create an answer
			answer, err := pc.CreateAnswer(nil)
			if err != nil {
				log.Println("Failed to create answer:", err)
				return
			}

			// Set local description and send it to remote peer
			if err := pc.SetLocalDescription(answer); err != nil {
				log.Println("Failed to set local description:", err)
				return
			}

			// Send SDP answer to remote peer through WebSocket
			if err := conn.WriteJSON(answer); err != nil {
				log.Println("Failed to send SDP answer:", err)
			}
		}
	}

	// Read messages from WebSocket
	for {
		var msg map[string]interface{}
		if err := conn.ReadJSON(&msg); err != nil {
			log.Println("Failed to read WebSocket message:", err)
			break
		}

		switch msg["type"] {
		case "offer":
			offer := webrtc.SessionDescription{
				Type: webrtc.SDPTypeOffer,
				SDP:  msg["sdp"].(string),
			}
			handleSDP(offer)

		case "answer":
			answer := webrtc.SessionDescription{
				Type: webrtc.SDPTypeAnswer,
				SDP:  msg["sdp"].(string),
			}
			handleSDP(answer)

		case "candidate":
			sdpMid := msg["sdpMid"].(string)
			sdpMLineIndex := msg["sdpMLineIndex"].(uint16)
			candidate := webrtc.ICECandidateInit{
				Candidate:     msg["candidate"].(string),
				SDPMid:        &sdpMid,
				SDPMLineIndex: &sdpMLineIndex,
			}
			// Add ICE candidate
			if err := pc.AddICECandidate(candidate); err != nil {
				log.Println("Failed to add ICE candidate:", err)
			}
		case "close":
			err := pc.Close()
			if err != nil {
				log.Println("close peer connection", err.Error())
			}
			ok := perrLock.TryLock()
			if !ok {
				sendMsg(conn, []byte("连接数已满"))
				return
			}
			pc = nil
			peerConnections[key] = nil
			perrLock.Unlock()
			conn.Close()
		}
	}
}

func closePeerConnections() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
	perrLock.TryLock()
	for _, pc := range peerConnections {
		if pc != nil {
			func() {
				defer func() {
					if err := recover(); err != nil {
						return
					}
				}()
				e := pc.Close()
				log.Println("close peer connection", e.Error())
			}()

		}
	}
}

func sendMsg(conn *websocket.Conn, msg []byte) {
	e := conn.ReadJSON(msg)
	if e != nil {
		log.Println("send msg error", e.Error())
	}
}

type RTPRecorder struct {
	Track *webrtc.TrackRemote
	File  *os.File
}

func (r *RTPRecorder) Start() {
	go func() {
		for {
			packet, _, err := r.Track.ReadRTP()
			if err != nil {
				log.Println("Failed to read RTP packet:", err)
				break
			}

			// Write the packet to the video file
			if _, err := r.File.Write(packet.Payload); err != nil {
				log.Println("Failed to write video packet:", err)
				break
			}
		}

		r.File.Close()
	}()
}
