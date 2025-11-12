package platform_server

import (
	"context"
	"runtime/debug"
	"time"

	loggerlib "github.com/Burmachine/MSA/lib/logger"
	platform_middleware "github.com/Burmachine/MSA/lib/middleware"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	"golang.org/x/time/rate"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func NewServerResiliencyInterceptors(log *loggerlib.Logger, cfg platform_middleware.ServerConfig) []grpc.UnaryServerInterceptor {
	var interceptors []grpc.UnaryServerInterceptor

	// 1. Panic recovery: всегда первый (должен обернуть все остальные)
	interceptors = append(interceptors, grpc_recovery.UnaryServerInterceptor(
		grpc_recovery.WithRecoveryHandler(func(p interface{}) (err error) {
			// Создаем фоновый контекст для логирования panic
			ctx := context.Background()
			log.Error(ctx, "panic recovered",
				"panic", p,
				"stack", string(debug.Stack()),
			)
			return status.Error(codes.Internal, "internal server error")
		}),
	))

	// 2. Timeout interceptor
	if cfg.Timeout.Enabled && cfg.Timeout.TimeoutMs > 0 {
		interceptors = append(interceptors, createTimeoutInterceptor(log, cfg.Timeout))
	}

	// 3. Rate limit interceptor
	if cfg.RateLimit.Enabled && cfg.RateLimit.ReqPerSec > 0 {
		interceptors = append(interceptors, createRateLimitInterceptor(log, cfg.RateLimit))
	}

	return interceptors
}

// createTimeoutInterceptor создает интерсептор для управления таймаутами
func createTimeoutInterceptor(log *loggerlib.Logger, cfg platform_middleware.ServerConfig_Timeout) grpc.UnaryServerInterceptor {
	// Кэшируем таймауты для конкретных путей
	pathTimeouts := make(map[string]time.Duration, len(cfg.Paths))
	for _, p := range cfg.Paths {
		if p.TimeoutMs > 0 {
			pathTimeouts[p.Path] = time.Duration(p.TimeoutMs) * time.Millisecond
		}
	}

	defaultTimeout := time.Duration(cfg.TimeoutMs) * time.Millisecond

	// Кэшируем ignore list для быстрой проверки
	ignoreSet := make(map[string]struct{}, len(cfg.Ignore))
	for _, path := range cfg.Ignore {
		if path != "" {
			ignoreSet[path] = struct{}{}
		}
	}

	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		// Проверяем blacklist
		if _, ignored := ignoreSet[info.FullMethod]; ignored {
			return handler(ctx, req)
		}

		// Выбираем таймаут: per-path или дефолтный
		timeout := defaultTimeout
		if pathTimeout, exists := pathTimeouts[info.FullMethod]; exists {
			timeout = pathTimeout
		}

		// Создаем контекст с таймаутом
		ctx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()

		// Логируем для отладки
		log.Debug(ctx, "timeout applied",
			"method", info.FullMethod,
			"timeout", timeout,
		)

		return handler(ctx, req)
	}
}

// createRateLimitInterceptor создает интерсептор для rate limiting
func createRateLimitInterceptor(log *loggerlib.Logger, cfg platform_middleware.ServerConfig_RateLimit) grpc.UnaryServerInterceptor {
	// Глобальный лимитер
	globalLimiter := rate.NewLimiter(rate.Limit(cfg.ReqPerSec), cfg.ReqPerSec)

	// Per-path лимитеры
	pathLimiters := make(map[string]*rate.Limiter, len(cfg.Paths))
	ctx := context.Background()
	for _, p := range cfg.Paths {
		if p.ReqPerSec > 0 {
			pathLimiters[p.Path] = rate.NewLimiter(rate.Limit(p.ReqPerSec), p.ReqPerSec)
			log.Info(ctx, "per-path rate limiter configured",
				"path", p.Path,
				"reqPerSec", p.ReqPerSec,
			)
		}
	}

	// Кэшируем ignore list
	ignoreSet := make(map[string]struct{}, len(cfg.Ignore))
	for _, path := range cfg.Ignore {
		if path != "" {
			ignoreSet[path] = struct{}{}
		}
	}

	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		// Проверяем blacklist
		if _, ignored := ignoreSet[info.FullMethod]; ignored {
			return handler(ctx, req)
		}

		// Выбираем лимитер: per-path или глобальный
		limiter := globalLimiter
		if pathLimiter, exists := pathLimiters[info.FullMethod]; exists {
			limiter = pathLimiter
		}

		// Проверяем лимит
		if !limiter.Allow() {
			log.Warn(ctx, "rate limit exceeded",
				"method", info.FullMethod,
			)
			return nil, status.Error(codes.ResourceExhausted, "rate limit exceeded")
		}

		return handler(ctx, req)
	}
}

// NewServerResiliencyStreamInterceptors для stream - только panic recovery
func NewServerResiliencyStreamInterceptors(log *loggerlib.Logger) []grpc.StreamServerInterceptor {
	return []grpc.StreamServerInterceptor{
		grpc_recovery.StreamServerInterceptor(
			grpc_recovery.WithRecoveryHandler(func(p interface{}) (err error) {
				ctx := context.Background()
				log.Error(ctx, "panic recovered in stream",
					"panic", p,
					"stack", string(debug.Stack()),
				)
				return status.Error(codes.Internal, "internal server error")
			}),
		),
	}
}
