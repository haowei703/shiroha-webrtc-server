package webrtc

import (
	"errors"
	"fmt"
	"github.com/asticode/go-astiav"
	"github.com/pion/rtp"
	"github.com/pion/rtp/codecs"
	"image/png"
	"log"
	"os"
	"sync"
	"time"
)

// VideoDecoder 视频解码器
type VideoDecoder struct {
	ctx         *astiav.CodecContext
	frameBuffer map[uint32][]byte // 用于存储未完成的视频帧
	mu          sync.Mutex        // 用于并发保护 frameBuffer
}

func NewVideoDecoder() (*VideoDecoder, error) {
	vd := &VideoDecoder{
		frameBuffer: make(map[uint32][]byte),
	}
	err := vd.initDecoder()
	return vd, err
}

// initDecoder initializes the VP8 videoCodec context
func (vd *VideoDecoder) initDecoder() error {
	// Initialize FFmpeg
	astiav.SetLogLevel(astiav.LogLevelDebug)

	// Find VP8 videoCodec
	videoCodec := astiav.FindDecoder(astiav.CodecIDVp8)
	if videoCodec == nil {
		return fmt.Errorf("unsupported codec")
	}

	// Allocate a codec context for the videoCodec
	codecContext := astiav.AllocCodecContext(videoCodec)
	if codecContext == nil {
		return fmt.Errorf("failed to allocate codec context")
	}

	// Open the codec
	if err := codecContext.Open(videoCodec, nil); err != nil {
		codecContext.Free()
		return fmt.Errorf("error opening codec: %w", err)
	}
	vd.ctx = codecContext
	return nil
}

// processRTPPacket 处理RTP包
func (vd *VideoDecoder) processRTPPacket(packet *rtp.Packet, codec string) ([]byte, int, int, error) {
	vd.mu.Lock()
	defer vd.mu.Unlock()

	log.Println("the number of packets is ", packet.SequenceNumber)

	var nalUnit []byte
	var err error
	switch codec {
	case "VP8":
		var vp8Packet codecs.VP8Packet
		nalUnit, err = vp8Packet.Unmarshal(packet.Payload)
		if err != nil {
			return nil, 0, 0, fmt.Errorf("failed to unmarshal vp8 packet: %v", err)
		}
	default:
		return nil, 0, 0, fmt.Errorf("unsupported codec: %s", codec)
	}

	// 将 NAL 单元数据添加到缓冲区中
	if buf, ok := vd.frameBuffer[packet.SSRC]; ok {
		vd.frameBuffer[packet.SSRC] = append(buf, nalUnit...)
	} else {
		vd.frameBuffer[packet.SSRC] = nalUnit
	}

	// 检查 RTP 包的 Marker 位，Marker 位为 1 表示这是一个完整帧的结束
	if packet.Marker {
		frame := vd.frameBuffer[packet.SSRC]
		delete(vd.frameBuffer, packet.SSRC)
		return vd.processVideoFrame(frame, codec)
	}
	// 视频帧不完整则返回nil
	return nil, 0, 0, nil
}

// processVideoFrame 使用astiav处理视频帧，解码为RGB格式
func (vd *VideoDecoder) processVideoFrame(input []byte, codec string) ([]byte, int, int, error) {
	if codec == "VP8" {
		return vd.decodeVp8FrameToRGBArray(input)
	}
	return nil, 0, 0, nil
}

// decodeVp8FrameToRGBArray VP8视频帧解码为RGB格式
func (vd *VideoDecoder) decodeVp8FrameToRGBArray(input []byte) ([]byte, int, int, error) {
	packet := astiav.AllocPacket()
	if packet == nil {
		log.Fatal("could not allocate packet")
	}

	defer packet.Free()
	if err := packet.FromData(input); err != nil {
		return nil, 0, 0, fmt.Errorf("error allocating packet: %w", err)
	}

	frame := astiav.AllocFrame()
	defer frame.Free()

	if err := vd.ctx.SendPacket(packet); err != nil {
		return nil, 0, 0, fmt.Errorf("error sending packet to decoder: %w", err)
	}

	if err := vd.ctx.ReceiveFrame(frame); err != nil {
		if errors.Is(err, astiav.ErrEagain) || errors.Is(err, astiav.ErrEof) {
			return nil, 0, 0, fmt.Errorf("frame not available yet")
		}
		return nil, 0, 0, fmt.Errorf("error receiving frame from decoder: %w", err)
	}

	// 创建一个用于RGB数据的Frame
	rgbFrame := astiav.AllocFrame()
	defer rgbFrame.Free()
	rgbFrame.SetPixelFormat(astiav.PixelFormatRgba)
	rgbFrame.SetWidth(frame.Width())
	rgbFrame.SetHeight(frame.Height())
	if err := rgbFrame.AllocBuffer(1); err != nil {
		return nil, 0, 0, fmt.Errorf("error allocating RGB frame: %w", err)
	}

	swsCtx, err := astiav.CreateSoftwareScaleContext(
		frame.Width(), frame.Height(), frame.PixelFormat(),
		frame.Width(), frame.Height(), rgbFrame.PixelFormat(),
		astiav.NewSoftwareScaleContextFlags(astiav.SoftwareScaleContextFlagBilinear),
	)

	if err != nil {
		return nil, 0, 0, fmt.Errorf("error creating software scaler context: %w", err)
	}
	defer swsCtx.Free()

	// 执行缩放
	if err := swsCtx.ScaleFrame(frame, rgbFrame); err != nil {
		return nil, 0, 0, fmt.Errorf("error scaling frame: %w", err)
	}

	data, err := rgbFrame.Data().Bytes(1)
	if len(data) == 0 || err != nil {
		log.Printf("Decoded RGB data size: %d bytes", len(data))
		return nil, 0, 0, fmt.Errorf("no RGB data found")
	}
	rgbData := make([]byte, len(data))
	copy(rgbData, data)
	log.Println("sent time is", time.Now())
	return rgbData, frame.Width(), frame.Height(), nil
}

func (vd *VideoDecoder) writeToFile(rgbFrame *astiav.Frame) error {
	img, err := rgbFrame.Data().GuessImageFormat()
	if err != nil {
		return fmt.Errorf("error guessing RGB frame format: %w", err)
	}

	err = rgbFrame.Data().ToImage(img)
	if err != nil {
		return fmt.Errorf("error converting RGB frame to image: %w", err)
	}
	directory := "debug/"
	fileName := fmt.Sprintf("%s%s.png", directory, time.Now().Format("2006-01-02_15-04-05"))
	dstFile, err := os.Create(fileName)
	if err != nil {
		log.Fatal(fmt.Errorf("main: creating %s failed: %w", "test.png", err))
	}
	defer dstFile.Close()

	if err = png.Encode(dstFile, img); err != nil {
		encodeErr := fmt.Errorf("main: encoding to png failed: %w", err)
		log.Fatal(encodeErr, err)
		return encodeErr
	}
	return nil
}
