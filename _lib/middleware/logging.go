package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type LoggingMiddleware struct {
	logger *zap.Logger
}

func NewLoggingMiddleware(logger *zap.Logger) *LoggingMiddleware {
	return &LoggingMiddleware{logger: logger}
}

func (m *LoggingMiddleware) Handle(c *gin.Context) {
	start := time.Now()
	path := c.Request.URL.Path

	c.Next()

	m.logger.Info("http_request",
		zap.String("method", c.Request.Method),
		zap.String("path", path),
		zap.Int("status", c.Writer.Status()),
		zap.Duration("duration", time.Since(start)),
	)
}
