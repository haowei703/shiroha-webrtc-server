package webrtc

import (
	"errors"
	"fmt"
	"github.com/asticode/go-astiav"
	"github.com/pion/rtp"
	"image/png"
	"log"
	"os"
	"sync"
	"time"
)

var unmarshallerMap = map[string]PacketUnmarshaller{
	"VP8":  &VP8PacketUnmarshaller{},
	"VP9":  &VP9PacketUnmarshaller{},
	"H264": &H264PacketUnmarshaller{},
	"H265": &H265PacketUnmarshaller{},
}

// VideoDecoder 视频解码器
type VideoDecoder struct {
	ctx         *astiav.CodecContext
	frameBuffer map[uint32][]byte // 用于存储未完成的视频帧
	mu          sync.Mutex        // 用于并发保护 frameBuffer
}

func NewVideoDecoder(codec astiav.CodecID) (*VideoDecoder, error) {
	vd := &VideoDecoder{
		frameBuffer: make(map[uint32][]byte),
	}
	err := vd.initDecoder(codec)
	return vd, err
}

// initDecoder initializes the VP8 videoCodec context
func (vd *VideoDecoder) initDecoder(codec astiav.CodecID) error {
	// Initialize FFmpeg
	astiav.SetLogLevel(astiav.LogLevelDebug)

	videoCodec := astiav.FindDecoder(codec)
	if videoCodec == nil {
		return fmt.Errorf("unsupported codec")
	}

	log.Printf("The decoder currently in use is %s\n", codec)

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

	var nalUnit []byte
	var err error

	unmarshaller, ok := unmarshallerMap[codec]
	if !ok {
		return nil, 0, 0, fmt.Errorf("unsupported codec: %s", codec)
	}
	nalUnit, err = unmarshaller.Unmarshal(packet.Payload)
	if err != nil {
		return nil, 0, 0, fmt.Errorf("error unmarshalling payload: %w", err)
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
		return vd.decodeFrameToRGBArray(frame)
	}
	// 视频帧不完整则返回nil
	return nil, 0, 0, nil
}

// decodeFrameToRGBArray 视频帧解码为RGB格式
func (vd *VideoDecoder) decodeFrameToRGBArray(input []byte) ([]byte, int, int, error) {
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
