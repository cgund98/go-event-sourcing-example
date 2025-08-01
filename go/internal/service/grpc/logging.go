package grpc

import (
	"context"
	"time"

	"github.com/cgund98/go-eventsrc-example/internal/infra/logging"
	"google.golang.org/grpc"
)

func LoggerInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	logging.Logger.Info("Received request", "grpc.method", info.FullMethod, "request.time", time.Now().Format(time.RFC3339))

	return handler(ctx, req)
}
