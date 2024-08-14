package webrtc

import (
	"github.com/pion/rtp"
	"github.com/pion/rtp/codecs"
)

type PacketUnmarshaller interface {
	Unmarshal(packet *rtp.Packet) ([]byte, error)
}

type VP8PacketUnmarshaller struct {
	frameBuffer map[uint32][]byte
	vp8Packet   *codecs.VP8Packet
}

func (u *VP8PacketUnmarshaller) Unmarshal(packet *rtp.Packet) ([]byte, error) {
	if u.vp8Packet == nil {
		u.vp8Packet = &codecs.VP8Packet{}
	}

	var nalUnit []byte

	// 将 NAL 单元数据添加到缓冲区中
	if buf, ok := u.frameBuffer[packet.SSRC]; ok {
		u.frameBuffer[packet.SSRC] = append(buf, nalUnit...)
	} else {
		u.frameBuffer[packet.SSRC] = nalUnit
	}

	// 检查 RTP 包的 Marker 位，Marker 位为 1 表示这是一个RTP序列的最后一个包
	if packet.Marker {
		frame := u.frameBuffer[packet.SSRC]
		delete(u.frameBuffer, packet.SSRC)
		return frame, nil
	}
	return u.vp8Packet.Unmarshal(packet.Payload)
}

type VP9PacketUnmarshaller struct {
	frameBuffer map[uint32][]byte
}

func (u *VP9PacketUnmarshaller) Unmarshal(packet *rtp.Packet) ([]byte, error) {
	var vp9Packet codecs.VP9Packet
	return vp9Packet.Unmarshal(packet.Payload)
}

type H264PacketUnmarshaller struct {
	frameBuffer map[uint32][]byte
	h264Packet  *codecs.H264Packet
}

func (u *H264PacketUnmarshaller) Unmarshal(packet *rtp.Packet) ([]byte, error) {
	if u.h264Packet == nil {
		u.h264Packet = &codecs.H264Packet{}
		u.frameBuffer = make(map[uint32][]byte)
	}
	nalUnit, err := u.h264Packet.Unmarshal(packet.Payload)
	if err != nil {
		return nil, err
	}
	if nalUnit == nil {
		return nil, nil
	}

	// 将 NAL 单元数据添加到缓冲区中
	if buf, ok := u.frameBuffer[packet.SSRC]; ok {
		u.frameBuffer[packet.SSRC] = append(buf, nalUnit...)
	} else {
		u.frameBuffer[packet.SSRC] = nalUnit
	}
	// 检查 RTP 包的 Marker 位，Marker 位为 1 表示这是一个RTP序列的最后一个包
	if packet.Marker {
		frame := u.frameBuffer[packet.SSRC]
		delete(u.frameBuffer, packet.SSRC)
		return frame, nil
	}
	return nil, nil
}

type H265PacketUnmarshaller struct{}

func (u *H265PacketUnmarshaller) Unmarshal(packet *rtp.Packet) ([]byte, error) {
	var h265Packet codecs.H265Packet
	return h265Packet.Unmarshal(packet.Payload)
}
