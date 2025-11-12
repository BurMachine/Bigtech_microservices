package platform_server

import (
	"context"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/Burmachine/MSA/lib/logger"
	platform_middleware "github.com/Burmachine/MSA/lib/middleware"
	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

func NewHTTPMiddlewares(log *loggerlib.Logger, cfg platform_middleware.ServerConfig) []gin.HandlerFunc {
	var middlewares []gin.HandlerFunc

	// 1. Panic recovery: всегда первый
	middlewares = append(middlewares, createPanicRecoveryMiddleware(log))

	// 2. Tracing (создает span для всей цепочки)
	middlewares = append(middlewares, createHTTPTracingMiddleware())

	// 3. Timeout middleware
	if cfg.Timeout.Enabled && cfg.Timeout.TimeoutMs > 0 {
		middlewares = append(middlewares, createHTTPTimeoutMiddleware(log, cfg.Timeout))
	}

	// 4. Rate limit middleware
	if cfg.RateLimit.Enabled && cfg.RateLimit.ReqPerSec > 0 {
		middlewares = append(middlewares, createHTTPRateLimitMiddleware(log, cfg.RateLimit))
	}

	// 5. Logging (последним - логирует финальный результат)
	middlewares = append(middlewares, createHTTPLoggingMiddleware(log))

	return middlewares
}

// createPanicRecoveryMiddleware перехватывает panic и возвращает 500
func createPanicRecoveryMiddleware(log *loggerlib.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// Логируем panic с полным stack trace
				log.Error(c.Request.Context(), "panic recovered",
					"panic", err,
					"path", c.Request.URL.Path,
					"method", c.Request.Method,
					"stack", string(debug.Stack()),
				)

				// Возвращаем 500 Internal Server Error
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
					"error": "internal server error",
				})
			}
		}()

		c.Next()
	}
}

// createHTTPTimeoutMiddleware создает middleware для управления таймаутами
func createHTTPTimeoutMiddleware(log *loggerlib.Logger, cfg platform_middleware.ServerConfig_Timeout) gin.HandlerFunc {
	// Кэшируем таймауты для конкретных путей
	pathTimeouts := make(map[string]time.Duration, len(cfg.Paths))
	for _, p := range cfg.Paths {
		if p.TimeoutMs > 0 {
			pathTimeouts[p.Path] = time.Duration(p.TimeoutMs) * time.Millisecond
		}
	}

	defaultTimeout := time.Duration(cfg.TimeoutMs) * time.Millisecond

	// Кэшируем ignore list
	ignoreSet := make(map[string]struct{}, len(cfg.Ignore))
	for _, path := range cfg.Ignore {
		if path != "" {
			ignoreSet[path] = struct{}{}
		}
	}

	return func(c *gin.Context) {
		// Формируем полный путь для проверки (метод + путь)
		fullPath := c.Request.Method + " " + c.Request.URL.Path

		// Проверяем blacklist
		if _, ignored := ignoreSet[fullPath]; ignored {
			c.Next()
			return
		}

		// Выбираем таймаут: per-path или дефолтный
		timeout := defaultTimeout
		if pathTimeout, exists := pathTimeouts[fullPath]; exists {
			timeout = pathTimeout
		}

		// Создаем контекст с таймаутом
		ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
		defer cancel()

		// Заменяем контекст запроса
		c.Request = c.Request.WithContext(ctx)

		// Логируем для отладки
		log.Debug(ctx, "timeout applied",
			"path", fullPath,
			"timeout", timeout,
		)

		// Канал для отслеживания завершения запроса
		finished := make(chan struct{})
		go func() {
			c.Next()
			close(finished)
		}()

		// Ждем либо завершения запроса, либо таймаута
		select {
		case <-finished:
			return
		case <-ctx.Done():
			log.Warn(ctx, "request timeout",
				"path", fullPath,
				"timeout", timeout,
			)
			c.AbortWithStatusJSON(http.StatusGatewayTimeout, gin.H{
				"error": "request timeout",
			})
		}
	}
}

// createHTTPRateLimitMiddleware создает middleware для rate limiting
func createHTTPRateLimitMiddleware(log *loggerlib.Logger, cfg platform_middleware.ServerConfig_RateLimit) gin.HandlerFunc {
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

	return func(c *gin.Context) {
		// Формируем полный путь для проверки
		fullPath := c.Request.Method + " " + c.Request.URL.Path

		// Проверяем blacklist
		if _, ignored := ignoreSet[fullPath]; ignored {
			c.Next()
			return
		}

		// Выбираем лимитер: per-path или глобальный
		limiter := globalLimiter
		if pathLimiter, exists := pathLimiters[fullPath]; exists {
			limiter = pathLimiter
		}

		// Проверяем лимит
		if !limiter.Allow() {
			log.Warn(c.Request.Context(), "rate limit exceeded",
				"path", fullPath,
			)
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": "rate limit exceeded",
			})
			return
		}

		c.Next()
	}
}
