package platform_server

import (
	"strconv"
	"time"

	"github.com/Burmachine/MSA/lib/metrics"
	"github.com/gin-gonic/gin"
)

// NewHTTPMetricsMiddleware создает middleware для HTTP метрик
func NewHTTPMetricsMiddleware(m *metrics.Metrics, serviceName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Извлекаем информацию о запросе
		method := c.Request.Method // GET, POST, etc
		path := c.FullPath()       // /v1/users/:id

		// Если path пустой (404), используем Request.URL.Path
		if path == "" {
			path = c.Request.URL.Path
		}

		// Увеличиваем счетчик активных запросов
		m.HTTPRequestsInFlight.WithLabelValues(serviceName, method, path).Inc()
		defer m.HTTPRequestsInFlight.WithLabelValues(serviceName, method, path).Dec()

		// Засекаем время начала
		start := time.Now()

		// Обрабатываем запрос
		c.Next()

		// Вычисляем длительность
		duration := time.Since(start)

		// Получаем HTTP status code
		statusCode := c.Writer.Status()
		code := strconv.Itoa(statusCode)

		// Записываем метрики
		m.RecordHTTPRequest(serviceName, method, path, code, duration)
	}
}
