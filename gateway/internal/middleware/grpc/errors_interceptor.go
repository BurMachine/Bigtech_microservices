package middleware_grpc

import (
	"context"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// LoggingUnaryServerInterceptor - логирование входящих запросов
func LoggingUnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		start := time.Now()

		log.Printf("[GATEWAY IN] → method=%s", info.FullMethod)

		resp, err := handler(ctx, req)

		duration := time.Since(start)
		if err != nil {
			st, _ := status.FromError(err)
			log.Printf("[GATEWAY IN] ← method=%s duration=%v code=%s error=%v",
				info.FullMethod, duration, st.Code(), err)
		} else {
			log.Printf("[GATEWAY IN] ← method=%s duration=%v code=OK",
				info.FullMethod, duration)
		}

		return resp, err
	}
}

// ErrorUnaryServerInterceptor - обработка ошибок от downstream сервисов
func ErrorUnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		resp, err := handler(ctx, req)
		if err != nil {
			// Если ошибка уже gRPC статус (пришла от downstream), пробрасываем дальше
			if st, ok := status.FromError(err); ok {
				// Можно добавить дополнительный контекст
				log.Printf("[GATEWAY ERROR] method=%s downstream_error=%s message=%s",
					info.FullMethod, st.Code(), st.Message())
				return resp, err
			}

			// Если это внутренняя ошибка gateway
			log.Printf("[GATEWAY ERROR] method=%s internal_error=%v", info.FullMethod, err)
			return nil, status.Error(codes.Internal, "gateway internal error")
		}
		return resp, nil
	}
}

// RecoveryUnaryServerInterceptor - защита от паник
func RecoveryUnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("[GATEWAY PANIC] method=%s panic=%v", info.FullMethod, r)
				err = status.Error(codes.Internal, "internal server error")
			}
		}()
		return handler(ctx, req)
	}
}
