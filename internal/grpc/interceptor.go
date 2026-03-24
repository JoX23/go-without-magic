package grpc

import (
	"context"
	"fmt"

	"go.uber.org/zap"
	grpcpkg "google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// UnaryServerInterceptor provee middleware reusable para manejar errores y panics.
func UnaryServerInterceptor(logger *zap.Logger) grpcpkg.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpcpkg.UnaryServerInfo, handler grpcpkg.UnaryHandler) (resp interface{}, err error) {
		defer func() {
			if r := recover(); r != nil {
				panicErr := fmt.Errorf("panic recovered: %v", r)
				logger.Error("grpc panic recovered", zap.Any("panic", r), zap.Error(panicErr))
				err = status.Error(codes.Internal, "internal server error")
			}
		}()

		resp, err = handler(ctx, req)
		if err != nil {
			mapped := ToGRPCError(err)
			logger.Error("grpc handler returned error", zap.String("method", info.FullMethod), zap.Error(err), zap.Error(mapped))
			return resp, mapped
		}

		return resp, nil
	}
}
