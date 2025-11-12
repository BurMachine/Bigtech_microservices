package platform_server

import (
	"context"
	"time"

	loggerlib "github.com/Burmachine/MSA/lib/logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
)

// NewServerLoggingInterceptors создает логирующие интерсепторы для gRPC сервера
func NewServerObservabilityInterceptors(log *loggerlib.Logger) []grpc.UnaryServerInterceptor {
	return []grpc.UnaryServerInterceptor{
		createLoggingInterceptor(log),
	}
}

// createLoggingInterceptor создает интерсептор для логирования запросов
func createLoggingInterceptor(log *loggerlib.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		start := time.Now()

		// Извлекаем метаданные
		requestID := extractRequestID(ctx)
		clientIP := extractClientIP(ctx)
		userAgent := extractUserAgent(ctx)

		// Добавляем request_id в контекст для последующих логов
		if requestID != "" {
			ctx = loggerlib.WithRequestID(ctx, requestID)
		}

		// Логируем начало запроса
		log.Debug(ctx, "gRPC request started",
			"method", info.FullMethod,
			"client_ip", clientIP,
			"user_agent", userAgent,
		)

		// Выполняем handler
		resp, err := handler(ctx, req)

		// Вычисляем длительность
		duration := time.Since(start)

		// Получаем gRPC status
		st, _ := status.FromError(err)
		code := st.Code()

		// Логируем результат
		if err != nil {
			// Определяем уровень логирования в зависимости от кода ошибки
			if isClientError(code) {
				// Клиентские ошибки (4xx) - info/warn
				log.Warn(ctx, "gRPC request completed with client error",
					"method", info.FullMethod,
					"duration_ms", duration.Milliseconds(),
					"code", code.String(),
					"error", st.Message(),
					"client_ip", clientIP,
				)
			} else {
				// Серверные ошибки (5xx) - error
				log.Error(ctx, "gRPC request failed",
					"method", info.FullMethod,
					"duration_ms", duration.Milliseconds(),
					"code", code.String(),
					"error", st.Message(),
					"client_ip", clientIP,
				)
			}
		} else {
			// Успешный запрос
			log.Info(ctx, "gRPC request completed",
				"method", info.FullMethod,
				"duration_ms", duration.Milliseconds(),
				"code", code.String(),
				"client_ip", clientIP,
			)
		}

		return resp, err
	}
}

// NewServerLoggingStreamInterceptors создает логирующие интерсепторы для stream
func NewServerObservabilityStreamInterceptors(log *loggerlib.Logger) []grpc.StreamServerInterceptor {
	return []grpc.StreamServerInterceptor{
		createStreamLoggingInterceptor(log),
	}
}

// createStreamLoggingInterceptor создает интерсептор для логирования stream запросов
func createStreamLoggingInterceptor(log *loggerlib.Logger) grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		start := time.Now()
		ctx := ss.Context()

		// Извлекаем метаданные
		requestID := extractRequestID(ctx)
		clientIP := extractClientIP(ctx)

		if requestID != "" {
			ctx = loggerlib.WithRequestID(ctx, requestID)
		}

		log.Debug(ctx, "gRPC stream started",
			"method", info.FullMethod,
			"client_ip", clientIP,
			"is_client_stream", info.IsClientStream,
			"is_server_stream", info.IsServerStream,
		)

		// Оборачиваем stream для передачи обновленного контекста
		wrappedStream := &wrappedServerStream{
			ServerStream: ss,
			ctx:          ctx,
		}

		// Выполняем handler
		err := handler(srv, wrappedStream)

		duration := time.Since(start)
		st, _ := status.FromError(err)
		code := st.Code()

		if err != nil {
			if isClientError(code) {
				log.Warn(ctx, "gRPC stream completed with client error",
					"method", info.FullMethod,
					"duration_ms", duration.Milliseconds(),
					"code", code.String(),
					"error", st.Message(),
				)
			} else {
				log.Error(ctx, "gRPC stream failed",
					"method", info.FullMethod,
					"duration_ms", duration.Milliseconds(),
					"code", code.String(),
					"error", st.Message(),
				)
			}
		} else {
			log.Info(ctx, "gRPC stream completed",
				"method", info.FullMethod,
				"duration_ms", duration.Milliseconds(),
				"code", code.String(),
			)
		}

		return err
	}
}

// wrappedServerStream оборачивает ServerStream для передачи обновленного контекста
type wrappedServerStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (w *wrappedServerStream) Context() context.Context {
	return w.ctx
}

// Helper функции

// extractRequestID извлекает request_id из metadata
func extractRequestID(ctx context.Context) string {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ""
	}

	// Проверяем различные варианты header
	headers := []string{"x-request-id", "request-id", "x-correlation-id"}
	for _, header := range headers {
		if values := md.Get(header); len(values) > 0 {
			return values[0]
		}
	}

	return ""
}

// extractClientIP извлекает IP клиента
func extractClientIP(ctx context.Context) string {
	// Сначала проверяем x-forwarded-for (если за прокси)
	md, ok := metadata.FromIncomingContext(ctx)
	if ok {
		if values := md.Get("x-forwarded-for"); len(values) > 0 {
			return values[0]
		}
		if values := md.Get("x-real-ip"); len(values) > 0 {
			return values[0]
		}
	}

	// Иначе берем из peer info
	p, ok := peer.FromContext(ctx)
	if ok {
		return p.Addr.String()
	}

	return "unknown"
}

// extractUserAgent извлекает user-agent
func extractUserAgent(ctx context.Context) string {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ""
	}

	if values := md.Get("user-agent"); len(values) > 0 {
		return values[0]
	}
	if values := md.Get("grpc-user-agent"); len(values) > 0 {
		return values[0]
	}

	return ""
}

// isClientError определяет, является ли ошибка клиентской (4xx)
func isClientError(code codes.Code) bool {
	switch code {
	case codes.InvalidArgument,
		codes.NotFound,
		codes.AlreadyExists,
		codes.PermissionDenied,
		codes.Unauthenticated,
		codes.FailedPrecondition,
		codes.OutOfRange:
		return true
	default:
		return false
	}
}
