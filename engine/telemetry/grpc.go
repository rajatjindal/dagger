package telemetry

import (
	"context"

	grpc "google.golang.org/grpc"
	"google.golang.org/protobuf/proto"

	"github.com/dagger/dagger/engine/slog"
)

func MeasuringUnaryClientInterceptor() grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		reqSize := proto.Size(req.(proto.Message))
		err := invoker(ctx, method, req, reply, cc, opts...)
		respSize := proto.Size(reply.(proto.Message))
		slog.Info("measuring gRPC client request",
			"reqSize", reqSize,
			"respSize", respSize)
		return err
	}
}

func MeasuringUnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		reqSize := proto.Size(req.(proto.Message))
		resp, err = handler(ctx, req)
		respSize := proto.Size(resp.(proto.Message))
		slog.Info("measuring gRPC server method",
			"method", info.FullMethod,
			"reqSize", reqSize,
			"respSize", respSize)
		return resp, err
	}
}

func MeasuringStreamClientInterceptor() grpc.StreamClientInterceptor {
	return func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
		clientStream, err := streamer(ctx, desc, cc, method, opts...)
		if err != nil {
			return nil, err
		}
		return &measuringClientStream{ClientStream: clientStream}, nil
	}
}

type measuringClientStream struct {
	grpc.ClientStream
}

func (s *measuringClientStream) SendMsg(m any) error {
	msgSize := proto.Size(m.(proto.Message))
	slog.Info("measuring client stream SendMsg", "msgSize", msgSize)
	return s.ClientStream.SendMsg(m)
}

func (s *measuringClientStream) RecvMsg(m any) error {
	err := s.ClientStream.RecvMsg(m)
	if err == nil {
		msgSize := proto.Size(m.(proto.Message))
		slog.Info("measuring client stream RecvMsg", "msgSize", msgSize)
	}
	return err
}
