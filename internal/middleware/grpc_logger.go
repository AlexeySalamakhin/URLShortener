package middleware

import (
	"context"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"

	"go.uber.org/zap"

	"github.com/AlexeySalamakhin/URLShortener/internal/logger"
)

// GRPCUnaryLogger логирует unary gRPC-запросы и ответы с длительностью и кодом статуса.
func GRPCUnaryLogger(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (interface{}, error) {
	start := time.Now()
	resp, err := handler(ctx, req)
	dur := time.Since(start)

	code := status.Code(err)
	remote := ""
	if p, ok := peer.FromContext(ctx); ok && p != nil && p.Addr != nil {
		remote = p.Addr.String()
	}

	logger.Log.Info("gRPC unary",
		zap.String("method", info.FullMethod),
		zap.Duration("duration", dur),
		zap.String("remote", remote),
		zap.String("code", code.String()),
	)
	return resp, err
}

// GRPCStreamLogger логирует stream gRPC-вызовы.
func GRPCStreamLogger(
	srv interface{},
	ss grpc.ServerStream,
	info *grpc.StreamServerInfo,
	handler grpc.StreamHandler,
) error {
	start := time.Now()
	ctx := ss.Context()
	remote := ""
	if p, ok := peer.FromContext(ctx); ok && p != nil && p.Addr != nil {
		remote = p.Addr.String()
	}

	err := handler(srv, ss)
	dur := time.Since(start)

	code := status.Code(err)
	if code == codes.OK && err == nil {
		code = codes.OK
	}

	logger.Log.Info("gRPC stream",
		zap.String("method", info.FullMethod),
		zap.Duration("duration", dur),
		zap.String("remote", remote),
		zap.String("code", code.String()),
	)
	return err
}
