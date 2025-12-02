// lib/middleware/server/http_auth_middleware.go

package platform_server

import (
	"fmt"
	"net/http"
	"strings"

	loggerlib "github.com/Burmachine/MSA/lib/logger"
	platform_middleware "github.com/Burmachine/MSA/lib/middleware"
	"github.com/gin-gonic/gin"
	"google.golang.org/grpc/metadata"
)

// ============================================
// HTTP AUTH MIDDLEWARE
// ============================================

type HTTPAuthMiddleware struct {
	authInterceptor *AuthInterceptor
	logger          *loggerlib.Logger
}

func NewHTTPAuthMiddleware(config *platform_middleware.AuthConfig, logger *loggerlib.Logger) (*HTTPAuthMiddleware, error) {
	// Создаём gRPC interceptor (вся логика валидации там)
	authInterceptor, err := NewAuthInterceptor(config, logger)
	if err != nil {
		return nil, err
	}

	return &HTTPAuthMiddleware{
		authInterceptor: authInterceptor,
		logger:          logger,
	}, nil
}

// Middleware возвращает Gin middleware функцию
func (m *HTTPAuthMiddleware) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		// 1. Проверяем публичные HTTP пути
		if m.isPublicPath(c.Request.URL.Path) {
			m.logger.Debug(ctx, "public path, skipping auth", "path", c.Request.URL.Path)
			c.Next()
			return
		}

		// 2. Извлекаем токен из Authorization header
		token, err := m.extractTokenFromHTTP(c)
		if err != nil {
			m.authInterceptor.metricsAuth.IncrementFailure()
			m.logger.Warn(ctx, "failed to extract token", "path", c.Request.URL.Path, "error", err)

			if m.authInterceptor.config.Required {
				c.JSON(http.StatusUnauthorized, gin.H{
					"error": "missing or invalid authorization header",
				})
				c.Abort()
				return
			}

			// Если auth.required = false, пропускаем
			c.Next()
			return
		}

		// 3. ✅ Валидируем JWT (переиспользуем логику из gRPC interceptor)
		authCtx, err := m.authInterceptor.validateJWT(ctx, token)
		if err != nil {
			m.authInterceptor.metricsAuth.IncrementFailure()
			m.logger.Warn(ctx, "JWT validation failed", "path", c.Request.URL.Path, "error", err)

			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "invalid or expired token",
			})
			c.Abort()
			return
		}

		// 4. Добавляем AuthContext в Gin context
		c.Set(authContextKey, authCtx)
		c.Set(userIDKey, authCtx.UserID)

		// 5. ✅ ВАЖНО: Добавляем токен в gRPC metadata для downstream вызовов
		md := metadata.Pairs(
			authorizationKey, "Bearer "+token,
			userIDKey, authCtx.UserID,
		)
		ctx = metadata.NewOutgoingContext(ctx, md)
		c.Request = c.Request.WithContext(ctx)

		m.authInterceptor.metricsAuth.IncrementSuccess()

		m.logger.Debug(ctx, "user authenticated via HTTP",
			"user_id", authCtx.UserID,
			"path", c.Request.URL.Path,
		)

		c.Next()
	}
}

// extractTokenFromHTTP извлекает Bearer токен из HTTP Authorization header
func (m *HTTPAuthMiddleware) extractTokenFromHTTP(c *gin.Context) (string, error) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		return "", fmt.Errorf("missing authorization header")
	}

	// Извлекаем Bearer токен
	token := strings.TrimPrefix(authHeader, "Bearer ")
	if token == authHeader {
		return "", fmt.Errorf("invalid authorization header format")
	}

	return token, nil
}

// isPublicPath проверяет, является ли HTTP путь публичным
func (m *HTTPAuthMiddleware) isPublicPath(path string) bool {
	if m.authInterceptor.config.Public.Paths == nil {
		return false
	}

	for _, publicPath := range m.authInterceptor.config.Public.Paths {
		// Точное совпадение или префикс
		if path == publicPath || strings.HasPrefix(path, publicPath) {
			return true
		}
	}
	return false
}

// ============================================
// HELPER FUNCTIONS для HTTP handlers
// ============================================

// GetAuthContextFromGin извлекает AuthContext из Gin context
func GetAuthContextFromGin(c *gin.Context) (*AuthContext, error) {
	value, exists := c.Get(authContextKey)
	if !exists {
		return nil, fmt.Errorf("auth context not found")
	}

	authCtx, ok := value.(*AuthContext)
	if !ok {
		return nil, fmt.Errorf("invalid auth context type")
	}

	return authCtx, nil
}

// GetUserIDFromGin извлекает user_id из Gin context
func GetUserIDFromGin(c *gin.Context) (string, error) {
	value, exists := c.Get(userIDKey)
	if !exists {
		return "", fmt.Errorf("user_id not found")
	}

	userID, ok := value.(string)
	if !ok {
		return "", fmt.Errorf("invalid user_id type")
	}

	return userID, nil
}

// MustGetUserIDFromGin извлекает user_id из Gin context (паника если нет)
func MustGetUserIDFromGin(c *gin.Context) string {
	userID, err := GetUserIDFromGin(c)
	if err != nil {
		panic(err)
	}
	return userID
}
