package grpc

import (
	"context"
	pb "github.com/haowei703/webrtc-server/github.com/haowei703/webrtc-server/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
	"os"
	"time"
)

const (
	defaultAddress = "localhost:50051" // The defaultAddress of the Python gRPC server
)

func SendMessage(videoFrame []byte, width int, height int) (string, error) {
	address := os.Getenv("GRPC_SERVER_ADDRESS")
	if address == "" {
		address = defaultAddress
	}
	conn, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer func(conn *grpc.ClientConn) {
		err := conn.Close()
		if err != nil {
			panic(err)
		}
	}(conn)
	c := pb.NewMessageExchangeClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	r, err := c.SendMessage(ctx, &pb.MessageRequest{VideoFrame: videoFrame, Width: int32(width), Height: int32(height)})
	if err != nil {
		return "", err
	}
	return r.GetResult(), nil
}
