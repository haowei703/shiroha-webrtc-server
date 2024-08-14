package webrtc

import "github.com/pion/rtp/codecs"

type PacketUnmarshaller interface {
	Unmarshal(payload []byte) ([]byte, error)
}

type VP8PacketUnmarshaller struct{}

func (u *VP8PacketUnmarshaller) Unmarshal(payload []byte) ([]byte, error) {
	var vp8Packet codecs.VP8Packet
	return vp8Packet.Unmarshal(payload)
}

type VP9PacketUnmarshaller struct{}

func (u *VP9PacketUnmarshaller) Unmarshal(payload []byte) ([]byte, error) {
	var vp9Packet codecs.VP9Packet
	return vp9Packet.Unmarshal(payload)
}

type H264PacketUnmarshaller struct{}

func (u *H264PacketUnmarshaller) Unmarshal(payload []byte) ([]byte, error) {
	var h264Packet codecs.H264Packet
	return h264Packet.Unmarshal(payload)
}

type H265PacketUnmarshaller struct{}

func (u *H265PacketUnmarshaller) Unmarshal(payload []byte) ([]byte, error) {
	var h265Packet codecs.H265Packet
	return h265Packet.Unmarshal(payload)
}
