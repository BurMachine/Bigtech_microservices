package platform_server

import (
	"context"
	"net/http"
	"runtime/debug"
	"time"

	platform_middleware "github.com/Burmachine/MSA/lib/middleware"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"golang.org/x/time/rate"
)

// NewHTTPMiddlewares создает цепочку HTTP middleware
func NewHTTPMiddlewares(logger *zap.Logger, cfg platform_middleware.ServerConfig) []gin.HandlerFunc {
	var middlewares []gin.HandlerFunc

	// 1. Panic recovery: всегда первый
	middlewares = append(middlewares, createPanicRecoveryMiddleware(logger))

	// 2. Timeout middleware
	if cfg.Timeout.Enabled && cfg.Timeout.TimeoutMs > 0 {
		middlewares = append(middlewares, createHTTPTimeoutMiddleware(logger, cfg.Timeout))
	}

	// 3. Rate limit middleware
	if cfg.RateLimit.Enabled && cfg.RateLimit.ReqPerSec > 0 {
		middlewares = append(middlewares, createHTTPRateLimitMiddleware(logger, cfg.RateLimit))
	}

	return middlewares
}

// createPanicRecoveryMiddleware перехватывает panic и возвращает 500
func createPanicRecoveryMiddleware(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// Логируем panic с полным stack trace
				logger.Error("panic recovered",
					zap.Any("panic", err),
					zap.String("path", c.Request.URL.Path),
					zap.String("method", c.Request.Method),
					zap.ByteString("stack", debug.Stack()),
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
func createHTTPTimeoutMiddleware(logger *zap.Logger, cfg platform_middleware.ServerConfig_Timeout) gin.HandlerFunc {
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
		logger.Debug("timeout applied",
			zap.String("path", fullPath),
			zap.Duration("timeout", timeout),
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
			// Запрос завершился нормально
			return
		case <-ctx.Done():
			// Таймаут истек
			logger.Warn("request timeout",
				zap.String("path", fullPath),
				zap.Duration("timeout", timeout),
			)
			c.AbortWithStatusJSON(http.StatusGatewayTimeout, gin.H{
				"error": "request timeout",
			})
		}
	}
}

// createHTTPRateLimitMiddleware создает middleware для rate limiting
func createHTTPRateLimitMiddleware(logger *zap.Logger, cfg platform_middleware.ServerConfig_RateLimit) gin.HandlerFunc {
	// Глобальный лимитер
	globalLimiter := rate.NewLimiter(rate.Limit(cfg.ReqPerSec), cfg.ReqPerSec)

	// Per-path лимитеры
	pathLimiters := make(map[string]*rate.Limiter, len(cfg.Paths))
	for _, p := range cfg.Paths {
		if p.ReqPerSec > 0 {
			pathLimiters[p.Path] = rate.NewLimiter(rate.Limit(p.ReqPerSec), p.ReqPerSec)
			logger.Info("per-path rate limiter configured",
				zap.String("path", p.Path),
				zap.Int("reqPerSec", p.ReqPerSec),
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
			logger.Warn("rate limit exceeded",
				zap.String("path", fullPath),
			)
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": "rate limit exceeded",
			})
			return
		}

		c.Next()
	}
}
