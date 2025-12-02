package platform_server

import (
	"context"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"strings"
	"sync"
	"time"

	loggerlib "github.com/Burmachine/MSA/lib/logger"
	"github.com/Burmachine/MSA/lib/middleware"
	"github.com/golang-jwt/jwt/v4"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// ============================================
// КОНСТАНТЫ
// ============================================

const (
	authorizationKey = "Authorization"
	userIDKey        = "user_id"
	authContextKey   = "auth_context"
)

// ============================================
// AUTH CONTEXT
// ============================================

// AuthContext содержит информацию об аутентифицированном пользователе
type AuthContext struct {
	UserID    string
	Issuer    string
	Audience  []string
	ExpiresAt time.Time
	IssuedAt  time.Time
	TokenID   string
	RawToken  string
}

// ============================================
// JWKS CLIENT (кеширование публичных ключей)
// ============================================

type JWKSClient struct {
	url       string
	cache     map[string]*rsa.PublicKey // kid -> RSA public key
	lastFetch time.Time
	cacheTTL  time.Duration
	timeout   time.Duration
	mutex     sync.RWMutex
	logger    *loggerlib.Logger
}

type JWK struct {
	KID string `json:"kid"`
	Kty string `json:"kty"`
	Use string `json:"use"`
	Alg string `json:"alg"`
	N   string `json:"n"`
	E   string `json:"e"`
}

type JWKSResponse struct {
	Keys []JWK `json:"keys"`
}

func NewJWKSClient(url string, cacheTTL, timeout time.Duration, logger *loggerlib.Logger) *JWKSClient {
	return &JWKSClient{
		url:      url,
		cache:    make(map[string]*rsa.PublicKey),
		cacheTTL: cacheTTL,
		timeout:  timeout,
		logger:   logger,
	}
}

// GetPublicKey получает публичный ключ по kid (с кешированием)
func (c *JWKSClient) GetPublicKey(ctx context.Context, kid string) (*rsa.PublicKey, error) {
	// 1. Проверяем кеш
	c.mutex.RLock()
	key, exists := c.cache[kid]
	needRefresh := time.Since(c.lastFetch) > c.cacheTTL
	c.mutex.RUnlock()

	if exists && !needRefresh {
		return key, nil
	}

	// 2. Обновляем кеш
	if err := c.RefreshCache(ctx); err != nil {
		c.logger.Error(ctx, "failed to refresh JWKS cache", "error", err)
		// Если обновление не удалось, но ключ в кеше - используем старый
		if exists {
			c.logger.Warn(ctx, "using stale cached key", "kid", kid)
			return key, nil
		}
		return nil, fmt.Errorf("failed to refresh JWKS cache: %w", err)
	}

	// 3. Пробуем снова из обновлённого кеша
	c.mutex.RLock()
	key, exists = c.cache[kid]
	c.mutex.RUnlock()

	if !exists {
		return nil, fmt.Errorf("kid not found in JWKS: %s", kid)
	}

	return key, nil
}

// RefreshCache обновляет кеш JWKS
func (c *JWKSClient) RefreshCache(ctx context.Context) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// Повторная проверка после получения lock
	if time.Since(c.lastFetch) < c.cacheTTL {
		return nil
	}

	c.logger.Debug(ctx, "refreshing JWKS cache", "url", c.url)

	// HTTP запрос с timeout
	httpCtx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(httpCtx, "GET", c.url, nil)
	if err != nil {
		return err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var jwksResp JWKSResponse
	if err := json.NewDecoder(resp.Body).Decode(&jwksResp); err != nil {
		return err
	}

	// Обновляем кеш
	newCache := make(map[string]*rsa.PublicKey)
	for _, jwk := range jwksResp.Keys {
		pubKey, err := jwkToRSAPublicKey(&jwk)
		if err != nil {
			c.logger.Warn(ctx, "failed to parse JWK", "kid", jwk.KID, "error", err)
			continue
		}
		newCache[jwk.KID] = pubKey
	}

	c.cache = newCache
	c.lastFetch = time.Now()

	c.logger.Info(ctx, "JWKS cache refreshed", "keys_count", len(newCache))
	return nil
}

// jwkToRSAPublicKey конвертирует JWK в RSA public key
func jwkToRSAPublicKey(jwk *JWK) (*rsa.PublicKey, error) {
	// Декодируем N (modulus)
	nBytes, err := base64.RawURLEncoding.DecodeString(jwk.N)
	if err != nil {
		return nil, fmt.Errorf("failed to decode N: %w", err)
	}

	// Декодируем E (exponent)
	eBytes, err := base64.RawURLEncoding.DecodeString(jwk.E)
	if err != nil {
		return nil, fmt.Errorf("failed to decode E: %w", err)
	}

	n := new(big.Int).SetBytes(nBytes)
	e := new(big.Int).SetBytes(eBytes)

	return &rsa.PublicKey{
		N: n,
		E: int(e.Int64()),
	}, nil
}

// ============================================
// AUTH INTERCEPTOR
// ============================================

type AuthInterceptor struct {
	config      *platform_middleware.AuthConfig
	jwksClient  *JWKSClient
	logger      *loggerlib.Logger
	metricsAuth *AuthMetrics
}

type AuthMetrics struct {
	successCount int64
	failureCount int64
	mutex        sync.Mutex
}

func NewAuthInterceptor(config *platform_middleware.AuthConfig, logger *loggerlib.Logger) (*AuthInterceptor, error) {
	if config == nil || !config.Enabled {
		return nil, fmt.Errorf("auth is not enabled")
	}

	jwksClient := NewJWKSClient(
		config.JWKS.URL,
		config.GetCacheTTL(),
		config.GetRefreshTimeout(),
		logger,
	)

	return &AuthInterceptor{
		config:      config,
		jwksClient:  jwksClient,
		logger:      logger,
		metricsAuth: &AuthMetrics{},
	}, nil
}

// UnaryInterceptor - gRPC unary interceptor для валидации JWT
func (a *AuthInterceptor) UnaryInterceptor(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (interface{}, error) {
	// 1. Проверяем публичные методы
	if a.config.IsPublicMethod(info.FullMethod) {
		a.logger.Debug(ctx, "public method, skipping auth", "method", info.FullMethod)
		return handler(ctx, req)
	}

	// 2. Извлекаем токен из metadata
	token, err := a.extractToken(ctx)
	if err != nil {
		a.metricsAuth.IncrementFailure()
		a.logger.Warn(ctx, "failed to extract token", "method", info.FullMethod, "error", err)
		if a.config.Required {
			return nil, status.Error(codes.Unauthenticated, "missing or invalid authorization header")
		}
		// Если auth.required = false, пропускаем
		return handler(ctx, req)
	}

	// 3. Валидируем JWT
	authCtx, err := a.validateJWT(ctx, token)
	if err != nil {
		a.metricsAuth.IncrementFailure()
		a.logger.Warn(ctx, "JWT validation failed", "method", info.FullMethod, "error", err)
		return nil, status.Error(codes.Unauthenticated, "invalid or expired token")
	}

	// 4. Добавляем AuthContext в context
	ctx = a.injectAuthContext(ctx, authCtx)
	a.metricsAuth.IncrementSuccess()

	a.logger.Debug(ctx, "user authenticated", "user_id", authCtx.UserID, "method", info.FullMethod)

	// 5. Вызываем handler
	return handler(ctx, req)
}

// extractToken извлекает Bearer токен из metadata
func (a *AuthInterceptor) extractToken(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", fmt.Errorf("missing metadata")
	}

	authHeader := md.Get(authorizationKey)
	if len(authHeader) == 0 {
		return "", fmt.Errorf("missing authorization header")
	}

	// Извлекаем Bearer токен
	token := strings.TrimPrefix(authHeader[0], "Bearer ")
	if token == authHeader[0] {
		return "", fmt.Errorf("invalid authorization header format")
	}

	return token, nil
}

// validateJWT валидирует JWT токен
func (a *AuthInterceptor) validateJWT(ctx context.Context, tokenString string) (*AuthContext, error) {
	// 1. Парсим токен без валидации (чтобы извлечь kid)
	token, _, err := new(jwt.Parser).ParseUnverified(tokenString, &jwt.RegisteredClaims{})
	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	// 2. Извлекаем kid из header
	kid, ok := token.Header["kid"].(string)
	if !ok {
		return nil, fmt.Errorf("missing kid in token header")
	}

	// 3. Получаем публичный ключ из JWKS
	publicKey, err := a.jwksClient.GetPublicKey(ctx, kid)
	if err != nil {
		return nil, fmt.Errorf("failed to get public key: %w", err)
	}

	// 4. Парсим и валидируем токен с публичным ключом
	parsedToken, err := jwt.ParseWithClaims(tokenString, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Проверяем алгоритм
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return publicKey, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to validate token: %w", err)
	}

	// 5. Извлекаем claims
	claims, ok := parsedToken.Claims.(*jwt.RegisteredClaims)
	if !ok || !parsedToken.Valid {
		return nil, fmt.Errorf("invalid token claims")
	}

	// 6. Проверяем issuer
	if claims.Issuer != a.config.Issuer {
		return nil, fmt.Errorf("invalid issuer: expected %s, got %s", a.config.Issuer, claims.Issuer)
	}

	// 7. Проверяем audience (если указан)
	if len(a.config.Audience) > 0 {
		validAudience := false
		for _, aud := range claims.Audience {
			for _, expectedAud := range a.config.Audience {
				if aud == expectedAud {
					validAudience = true
					break
				}
			}
		}
		if !validAudience {
			return nil, fmt.Errorf("invalid audience")
		}
	}

	// 8. Формируем AuthContext
	authCtx := &AuthContext{
		UserID:    claims.Subject,
		Issuer:    claims.Issuer,
		Audience:  claims.Audience,
		ExpiresAt: claims.ExpiresAt.Time,
		IssuedAt:  claims.IssuedAt.Time,
		TokenID:   claims.ID,
		RawToken:  tokenString,
	}

	return authCtx, nil
}

// injectAuthContext добавляет AuthContext в context
func (a *AuthInterceptor) injectAuthContext(ctx context.Context, authCtx *AuthContext) context.Context {
	ctx = context.WithValue(ctx, authContextKey, authCtx)
	ctx = context.WithValue(ctx, userIDKey, authCtx.UserID)
	return ctx
}

// ============================================
// HELPER FUNCTIONS
// ============================================

// GetAuthContext извлекает AuthContext из context
func GetAuthContext(ctx context.Context) (*AuthContext, error) {
	authCtx, ok := ctx.Value(authContextKey).(*AuthContext)
	if !ok {
		return nil, fmt.Errorf("auth context not found")
	}
	return authCtx, nil
}

// GetUserID извлекает user_id из context
func GetUserID(ctx context.Context) (string, error) {
	userID, ok := ctx.Value(userIDKey).(string)
	if !ok {
		return "", fmt.Errorf("user_id not found in context")
	}
	return userID, nil
}

// MustGetUserID извлекает user_id из context (паника если нет)
func MustGetUserID(ctx context.Context) string {
	userID, err := GetUserID(ctx)
	if err != nil {
		panic(err)
	}
	return userID
}

// ============================================
// METRICS
// ============================================

func (m *AuthMetrics) IncrementSuccess() {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.successCount++
}

func (m *AuthMetrics) IncrementFailure() {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.failureCount++
}

func (m *AuthMetrics) GetMetrics() (success, failure int64) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	return m.successCount, m.failureCount
}
