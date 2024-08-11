// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.4.0
// - protoc             v3.20.3
// source: proto/message.proto

package proto

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.62.0 or later.
const _ = grpc.SupportPackageIsVersion8

const (
	MessageExchange_SendMessage_FullMethodName = "/message.MessageExchange/SendMessage"
)

// MessageExchangeClient is the client API for MessageExchange service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type MessageExchangeClient interface {
	SendMessage(ctx context.Context, in *MessageRequest, opts ...grpc.CallOption) (*MessageResponse, error)
}

type messageExchangeClient struct {
	cc grpc.ClientConnInterface
}

func NewMessageExchangeClient(cc grpc.ClientConnInterface) MessageExchangeClient {
	return &messageExchangeClient{cc}
}

func (c *messageExchangeClient) SendMessage(ctx context.Context, in *MessageRequest, opts ...grpc.CallOption) (*MessageResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(MessageResponse)
	err := c.cc.Invoke(ctx, MessageExchange_SendMessage_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// MessageExchangeServer is the server API for MessageExchange service.
// All implementations must embed UnimplementedMessageExchangeServer
// for forward compatibility
type MessageExchangeServer interface {
	SendMessage(context.Context, *MessageRequest) (*MessageResponse, error)
	mustEmbedUnimplementedMessageExchangeServer()
}

// UnimplementedMessageExchangeServer must be embedded to have forward compatible implementations.
type UnimplementedMessageExchangeServer struct {
}

func (UnimplementedMessageExchangeServer) SendMessage(context.Context, *MessageRequest) (*MessageResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SendMessage not implemented")
}
func (UnimplementedMessageExchangeServer) mustEmbedUnimplementedMessageExchangeServer() {}

// UnsafeMessageExchangeServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to MessageExchangeServer will
// result in compilation errors.
type UnsafeMessageExchangeServer interface {
	mustEmbedUnimplementedMessageExchangeServer()
}

func RegisterMessageExchangeServer(s grpc.ServiceRegistrar, srv MessageExchangeServer) {
	s.RegisterService(&MessageExchange_ServiceDesc, srv)
}

func _MessageExchange_SendMessage_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MessageRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MessageExchangeServer).SendMessage(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: MessageExchange_SendMessage_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MessageExchangeServer).SendMessage(ctx, req.(*MessageRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// MessageExchange_ServiceDesc is the grpc.ServiceDesc for MessageExchange service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var MessageExchange_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "message.MessageExchange",
	HandlerType: (*MessageExchangeServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "SendMessage",
			Handler:    _MessageExchange_SendMessage_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "proto/message.proto",
}
