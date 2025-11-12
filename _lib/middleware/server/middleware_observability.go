package platform_server

import (
	"time"

	loggerlib "github.com/Burmachine/MSA/lib/logger"
	"github.com/gin-gonic/gin"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/uber/jaeger-client-go"
)

func createHTTPTracingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Определяем имя операции
		operationName := c.Request.Method + " " + c.FullPath()
		if c.FullPath() == "" {
			operationName = c.Request.Method + " " + c.Request.URL.Path
		}

		// ВСЕГДА создаем новый span (не проверяем контекст)
		span, ctx := opentracing.StartSpanFromContext(c.Request.Context(), operationName)
		defer span.Finish()

		// Обновляем контекст
		c.Request = c.Request.WithContext(ctx)

		// Извлекаем trace_id и добавляем в response
		spanContext, ok := span.Context().(jaeger.SpanContext)
		if ok {
			traceID := spanContext.TraceID().String()
			c.Header(traceIDKey, traceID)
		}

		// Теги
		ext.SpanKindRPCServer.Set(span)
		ext.Component.Set(span, "http")
		ext.HTTPMethod.Set(span, c.Request.Method)
		ext.HTTPUrl.Set(span, c.Request.URL.String())
		span.SetTag("http.route", c.FullPath())
		span.SetTag("http.client_ip", c.ClientIP())

		// Выполняем handler
		c.Next()

		// Результат
		statusCode := c.Writer.Status()
		ext.HTTPStatusCode.Set(span, uint16(statusCode))

		if statusCode >= 400 {
			ext.Error.Set(span, true)
			if len(c.Errors) > 0 {
				span.LogKV("http_error", c.Errors.String())
			}
		}

		span.SetTag("http.status_code", statusCode)
	}
}

func createHTTPLoggingMiddleware(log *loggerlib.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// Извлекаем метаданные
		requestID := c.GetHeader("x-request-id")
		if requestID == "" {
			requestID = c.GetHeader("request-id")
		}
		if requestID == "" {
			requestID = c.GetHeader("x-correlation-id")
		}

		clientIP := c.ClientIP()
		userAgent := c.GetHeader("user-agent")

		// Добавляем request_id в контекст
		ctx := c.Request.Context()
		if requestID != "" {
			ctx = loggerlib.WithRequestID(ctx, requestID)
			c.Request = c.Request.WithContext(ctx)
		}

		// Логируем начало запроса
		log.Debug(ctx, "HTTP request started",
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"client_ip", clientIP,
			"user_agent", userAgent,
		)

		// Выполняем handler
		c.Next()

		// Вычисляем длительность
		duration := time.Since(start)
		statusCode := c.Writer.Status()

		// Логируем результат
		if statusCode >= 500 {
			// Серверные ошибки (5xx)
			log.Error(ctx, "HTTP request failed",
				"method", c.Request.Method,
				"path", c.Request.URL.Path,
				"status", statusCode,
				"duration_ms", duration.Milliseconds(),
				"client_ip", clientIP,
				"errors", c.Errors.String(),
			)
		} else if statusCode >= 400 {
			// Клиентские ошибки (4xx)
			log.Warn(ctx, "HTTP request completed with client error",
				"method", c.Request.Method,
				"path", c.Request.URL.Path,
				"status", statusCode,
				"duration_ms", duration.Milliseconds(),
				"client_ip", clientIP,
			)
		} else {
			// Успешный запрос
			log.Info(ctx, "HTTP request completed",
				"method", c.Request.Method,
				"path", c.Request.URL.Path,
				"status", statusCode,
				"duration_ms", duration.Milliseconds(),
				"client_ip", clientIP,
			)
		}
	}
}
