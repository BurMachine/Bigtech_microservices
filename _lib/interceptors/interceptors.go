package interceptors

import (
	"context"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
)

type LoggingInterceptor struct {
	logger *zap.Logger
}

func NewLoggingInterceptor(logger *zap.Logger) *LoggingInterceptor {
	return &LoggingInterceptor{
		logger: logger,
	}
}

// UnaryServerInterceptor логирует унарные gRPC запросы
func (i *LoggingInterceptor) UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		start := time.Now()

		// Получаем IP клиента
		clientIP := "unknown"
		if p, ok := peer.FromContext(ctx); ok {
			clientIP = p.Addr.String()
		}

		// Вызываем обработчик
		resp, err := handler(ctx, req)

		// Определяем код ответа
		duration := time.Since(start)
		code := codes.OK
		errMsg := ""
		if err != nil {
			if st, ok := status.FromError(err); ok {
				code = st.Code()
				errMsg = st.Message()
			}
		}

		// Компактный JSON лог
		if err != nil {
			i.logger.Error("grpc_request",
				zap.String("method", info.FullMethod),
				zap.String("remote_addr", clientIP),
				zap.Duration("duration", duration),
				zap.String("status", code.String()),
				zap.String("error", errMsg),
			)
		} else {
			i.logger.Info("grpc_request",
				zap.String("method", info.FullMethod),
				zap.String("remote_addr", clientIP),
				zap.Duration("duration", duration),
				zap.String("status", code.String()),
			)
		}

		return resp, err
	}
}

// StreamServerInterceptor логирует потоковые gRPC запросы
func (i *LoggingInterceptor) StreamServerInterceptor() grpc.StreamServerInterceptor {
	return func(
		srv interface{},
		ss grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {
		start := time.Now()

		// Получаем IP клиента
		clientIP := "unknown"
		if p, ok := peer.FromContext(ss.Context()); ok {
			clientIP = p.Addr.String()
		}

		err := handler(srv, ss)

		duration := time.Since(start)
		code := codes.OK
		errMsg := ""
		if err != nil {
			if st, ok := status.FromError(err); ok {
				code = st.Code()
				errMsg = st.Message()
			}
		}

		if err != nil {
			i.logger.Error("grpc_stream",
				zap.String("method", info.FullMethod),
				zap.String("remote_addr", clientIP),
				zap.Duration("duration", duration),
				zap.String("status", code.String()),
				zap.String("error", errMsg),
			)
		} else {
			i.logger.Info("grpc_stream",
				zap.String("method", info.FullMethod),
				zap.String("remote_addr", clientIP),
				zap.Duration("duration", duration),
				zap.String("status", code.String()),
			)
		}

		return err
	}
}
