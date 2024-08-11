package webrtc

import (
	"encoding/json"
	"github.com/gorilla/websocket"
	"github.com/haowei703/webrtc-server/internal/grpc"
	"github.com/pion/webrtc/v3"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
)

type Message struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Failed to upgrade:", err)
		return
	}
	defer conn.Close()

	manager, err := NewWebRTCManager()
	if err != nil {
		log.Println("Failed to create RtcManager:", err)
		return
	}
	defer manager.Close()

	var mu sync.Mutex // 互斥锁，用于保护 WebSocket 连接

	writeMessage := func(messageType int, data []byte) error {
		mu.Lock()
		defer mu.Unlock()
		return conn.WriteMessage(messageType, data)
	}

	manager.PeerConnection.OnICECandidate(func(candidate *webrtc.ICECandidate) {
		if candidate != nil {
			jsonCandidate, _ := json.Marshal(candidate.ToJSON())
			message := Message{Type: "candidate", Data: jsonCandidate}
			jsonMessage, _ := json.Marshal(message)
			if err := writeMessage(websocket.TextMessage, jsonMessage); err != nil {
				log.Println("Failed to send ICE candidate:", err)
			}
		}
	})

	manager.PeerConnection.OnTrack(func(track *webrtc.TrackRemote, receiver *webrtc.RTPReceiver) {
		log.Printf("Got remote track: %s, type: %s\n", track.ID(), track.Kind())
		switch track.Kind() {
		case webrtc.RTPCodecTypeAudio:
			handleAudioTrack(track)
		case webrtc.RTPCodecTypeVideo:
			handleVideoTrack(track, writeMessage)
		}

		if err != nil {
			log.Println("Failed to init video decoder:", err)
			return
		}

	})

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Println("Read error:", err)
			break
		}

		var msg Message
		if err := json.Unmarshal(message, &msg); err != nil {
			log.Println("Unmarshal error:", err)
			log.Printf("Message: %s", message) // 打印原始消息进行调试
			continue
		}

		switch msg.Type {
		case "offer":
			var offer webrtc.SessionDescription
			if err := json.Unmarshal(msg.Data, &offer); err != nil {
				log.Println("Unmarshal offer error:", err)
				log.Printf("Data: %s", msg.Data) // 打印原始数据进行调试
				continue
			}
			answer, err := manager.HandleOffer(offer)
			if err != nil {
				log.Println("HandleOffer error:", err)
				continue
			}
			jsonAnswer, _ := json.Marshal(answer)
			responseMsg := Message{Type: "answer", Data: json.RawMessage(jsonAnswer)}
			jsonResponseMsg, _ := json.Marshal(responseMsg)
			if err := writeMessage(websocket.TextMessage, jsonResponseMsg); err != nil {
				log.Println("Failed to send answer:", err)
			}
		case "candidate":
			var candidate webrtc.ICECandidateInit
			if err := json.Unmarshal(msg.Data, &candidate); err != nil {
				log.Println("Unmarshal candidate error:", err)
				log.Printf("Data: %s", msg.Data) // 打印原始数据进行调试
				continue
			}
			if err := manager.AddICECandidate(candidate); err != nil {
				log.Println("AddICECandidate error:", err)
			}
		}
	}
}

func handleAudioTrack(track *webrtc.TrackRemote) {

}

func handleVideoTrack(track *webrtc.TrackRemote, writeMessage func(messageType int, data []byte) error) {
	vd, err := NewVideoDecoder()
	if err != nil {
		panic(err)
	}

	// 处理track
	for {
		rtp, _, readErr := track.ReadRTP()
		if readErr != nil {
			log.Println("ReadRTP error:", readErr)
			return
		}

		var rgbData []byte
		var width, height int

		mimeType := track.Codec().MimeType
		codec := strings.Split(mimeType, "/")[1]
		rgbData, width, height, err = vd.processRTPPacket(rtp, codec)
		if err != nil {
			log.Println("error processing RTP packet:", err)
		}

		// 视频帧不完整时退出当前循环继续处理
		if rgbData != nil {
			// 通过grpc将视频字节传输给下游
			response, err := grpc.SendMessage(rgbData, width, height)
			if err != nil {
				log.Println("gRPC error:", err)
				return
			}

			data := map[string]string{"message": response}
			jsonData, _ := json.Marshal(data)
			// 将处理结果回传给客户端
			msg := Message{Type: "text", Data: json.RawMessage(jsonData)}
			jsonMsg, _ := json.Marshal(msg)
			if err := writeMessage(websocket.TextMessage, jsonMsg); err != nil {
				log.Println("Failed to send response:", err)
			}
		}
	}
}

func StartWebSocketServer() {
	http.HandleFunc("/ws/signaling", handleWebSocket)
	port := os.Getenv("SIGNALING_PORT")
	if port == "" {
		port = "8081"
	}
	log.Printf("WebSocket server started at %s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
