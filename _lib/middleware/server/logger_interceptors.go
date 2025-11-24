package platform_server

import (
	"context"
	"encoding/json"
	"strings"
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

func createLoggingInterceptor(log *loggerlib.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		start := time.Now()

		requestID := extractRequestID(ctx)
		clientIP := extractClientIP(ctx)
		userAgent := extractUserAgent(ctx)

		// Проверяем debug режим из заголовка
		debugEnabled := isDebugEnabled(ctx)

		if requestID != "" {
			ctx = loggerlib.WithRequestID(ctx, requestID)
		}

		if debugEnabled {
			ctx = loggerlib.WithDebug(ctx, true)
		}

		// Логируем начало запроса
		logArgs := []interface{}{
			"method", info.FullMethod,
			"client_ip", clientIP,
			"user_agent", userAgent,
		}

		// Если debug режим - добавляем тело запроса и логируем на уровне INFO
		if debugEnabled {
			if reqJSON, err := json.Marshal(req); err == nil {
				logArgs = append(logArgs, "request_body", string(reqJSON))
			}
			log.Info(ctx, "gRPC request started (debug mode)", logArgs...)
		} else if log.IsDebugEnabled() {
			// Если глобальный DEBUG уровень - логируем на DEBUG
			if reqJSON, err := json.Marshal(req); err == nil {
				logArgs = append(logArgs, "request_body", string(reqJSON))
			}
			log.Debug(ctx, "gRPC request started (debug mode)", logArgs...)
		} else {
			// Обычное логирование
			log.Debug(ctx, "gRPC request started", logArgs...)
		}

		// Выполняем handler
		resp, err := handler(ctx, req)

		duration := time.Since(start)
		st, _ := status.FromError(err)
		code := st.Code()

		// Логируем результат
		resultArgs := []interface{}{
			"method", info.FullMethod,
			"duration_ms", duration.Milliseconds(),
			"code", code.String(),
			"client_ip", clientIP,
		}

		// Если debug режим и нет ошибки - добавляем тело ответа
		if debugEnabled && err == nil {
			if respJSON, marshalErr := json.Marshal(resp); marshalErr == nil {
				resultArgs = append(resultArgs, "response_body", string(respJSON))
			}
		} else if log.IsDebugEnabled() && err == nil {
			if respJSON, marshalErr := json.Marshal(resp); marshalErr == nil {
				resultArgs = append(resultArgs, "response_body", string(respJSON))
			}
		}

		if err != nil {
			resultArgs = append(resultArgs, "error", st.Message())
			if isClientError(code) {
				log.Warn(ctx, "gRPC request completed with client error", resultArgs...)
			} else {
				log.Error(ctx, "gRPC request failed", resultArgs...)
			}
		} else {
			// Если debug режим включен через заголовок - логируем на INFO
			if debugEnabled {
				log.Info(ctx, "gRPC request completed (debug mode)", resultArgs...)
			} else {
				log.Debug(ctx, "gRPC request completed", resultArgs...)
			}
		}

		return resp, err
	}
}

func isDebugEnabled(ctx context.Context) bool {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return false
	}

	// Проверяем различные варианты debug заголовков
	debugHeaders := []string{"x-debug", "x-debug-mode", "debug"}
	for _, header := range debugHeaders {
		if values := md.Get(header); len(values) > 0 {
			// Проверяем значение: "true", "1", "yes"
			val := strings.ToLower(values[0])
			if val == "true" || val == "1" || val == "yes" {
				return true
			}
		}
	}
	return false
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
		debugEnabled := isDebugEnabled(ctx)

		if requestID != "" {
			ctx = loggerlib.WithRequestID(ctx, requestID)
		}

		if debugEnabled {
			ctx = loggerlib.WithDebug(ctx, true)
		}

		// Логируем на INFO если debug режим, иначе на DEBUG
		if debugEnabled {
			log.Info(ctx, "gRPC stream started (debug mode)",
				"method", info.FullMethod,
				"client_ip", clientIP,
				"is_client_stream", info.IsClientStream,
				"is_server_stream", info.IsServerStream,
			)
		} else {
			log.Debug(ctx, "gRPC stream started",
				"method", info.FullMethod,
				"client_ip", clientIP,
				"is_client_stream", info.IsClientStream,
				"is_server_stream", info.IsServerStream,
			)
		}

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
			if debugEnabled {
				log.Info(ctx, "gRPC stream completed (debug mode)",
					"method", info.FullMethod,
					"duration_ms", duration.Milliseconds(),
					"code", code.String(),
				)
			} else {
				log.Info(ctx, "gRPC stream completed",
					"method", info.FullMethod,
					"duration_ms", duration.Milliseconds(),
					"code", code.String(),
				)
			}
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
