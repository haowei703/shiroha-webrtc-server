package webrtc

import (
	"encoding/json"
	"fmt"
	"github.com/pion/sdp/v3"
	"github.com/pion/webrtc/v3"
	"log"
)

// RtcManager webRTC连接管理类
type RtcManager struct {
	PeerConnection *webrtc.PeerConnection
}

func NewWebRTCManager() (*RtcManager, error) {
	// 创建 PeerConnection 配置
	config := webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{
				URLs: []string{"stun:stun.l.google.com:19302"},
			},
		},
	}

	// 创建新的 PeerConnection
	peerConnection, err := webrtc.NewPeerConnection(config)
	if err != nil {
		return nil, err
	}

	manager := &RtcManager{
		PeerConnection: peerConnection,
	}

	// 设置 ICE 候选者处理程序
	peerConnection.OnICECandidate(func(candidate *webrtc.ICECandidate) {
		if candidate != nil {
			jsonCandidate, _ := json.Marshal(candidate.ToJSON())
			// 这里可以将 ICE 候选者发送到远端
			log.Printf("New ICE Candidate: %s\n", jsonCandidate)
		}
	})

	peerConnection.OnICEConnectionStateChange(func(state webrtc.ICEConnectionState) {
		fmt.Printf("Connection State has changed %s \n", state.String())
	})

	// Create a video track
	// 使用vp8编码器
	videoTrack, err := webrtc.NewTrackLocalStaticSample(webrtc.RTPCodecCapability{MimeType: "video/vp8"}, "video", "pion")
	if err != nil {
		fmt.Printf("Error creating video track: %s\n", err)
	}

	_, err = peerConnection.AddTrack(videoTrack)
	if err != nil {
		fmt.Printf("Error creating video track: %s\n", err)
	}

	return manager, nil
}

func (manager *RtcManager) HandleOffer(offer webrtc.SessionDescription) (*webrtc.SessionDescription, error) {
	// 设置远端描述
	err := manager.PeerConnection.SetRemoteDescription(offer)
	if err != nil {
		return nil, err
	}

	// 创建应答
	answer, err := manager.PeerConnection.CreateAnswer(nil)
	if err != nil {
		return nil, err
	}

	// 设置本地描述
	err = manager.PeerConnection.SetLocalDescription(answer)
	if err != nil {
		return nil, err
	}

	parsedSDP := sdp.SessionDescription{}
	if err := parsedSDP.Unmarshal([]byte(offer.SDP)); err != nil {
		log.Fatalf("Failed to unmarshal SDP: %v", err)
	}

	return &answer, nil
}

func (manager *RtcManager) AddICECandidate(candidate webrtc.ICECandidateInit) error {
	return manager.PeerConnection.AddICECandidate(candidate)
}

func (manager *RtcManager) Close() error {
	return manager.PeerConnection.Close()
}
