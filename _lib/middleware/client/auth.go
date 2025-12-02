package platform_client

import (
	"context"
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// ============================================
// CLIENT AUTH INTERCEPTOR
// ============================================

// AuthClientInterceptor добавляет JWT токен в исходящие gRPC запросы
type AuthClientInterceptor struct {
	// Можно добавить конфигурацию если нужно
}

// NewClientAuthInterceptor создаёт interceptor для добавления токена в gRPC metadata
func NewClientAuthInterceptor() grpc.UnaryClientInterceptor {
	return func(
		ctx context.Context,
		method string,
		req, reply interface{},
		cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker,
		opts ...grpc.CallOption,
	) error {
		// Извлекаем токен из входящего context (если есть)
		token, err := extractTokenFromContext(ctx)
		if err != nil {
			// Если токена нет - продолжаем без него (может быть публичный метод)
			return invoker(ctx, method, req, reply, cc, opts...)
		}

		// Добавляем токен в исходящие metadata
		ctx = addTokenToOutgoingContext(ctx, token)

		// Вызываем оригинальный метод с обновлённым context
		return invoker(ctx, method, req, reply, cc, opts...)
	}
}

// NewClientAuthStreamInterceptor для stream запросов
func NewClientAuthStreamInterceptor() grpc.StreamClientInterceptor {
	return func(
		ctx context.Context,
		desc *grpc.StreamDesc,
		cc *grpc.ClientConn,
		method string,
		streamer grpc.Streamer,
		opts ...grpc.CallOption,
	) (grpc.ClientStream, error) {
		// Извлекаем токен из входящего context (если есть)
		token, err := extractTokenFromContext(ctx)
		if err == nil {
			// Добавляем токен в исходящие metadata
			ctx = addTokenToOutgoingContext(ctx, token)
		}

		// Вызываем оригинальный streamer
		return streamer(ctx, desc, cc, method, opts...)
	}
}

// extractTokenFromContext извлекает токен из входящих metadata
func extractTokenFromContext(ctx context.Context) (string, error) {
	// Пробуем извлечь из incoming metadata (если это HTTP → gRPC вызов)
	md, ok := metadata.FromIncomingContext(ctx)
	if ok {
		if authHeaders := md.Get("authorization"); len(authHeaders) > 0 {
			return authHeaders[0], nil
		}
	}

	// Пробуем извлечь из outgoing metadata (если уже был добавлен ранее)
	md, ok = metadata.FromOutgoingContext(ctx)
	if ok {
		if authHeaders := md.Get("authorization"); len(authHeaders) > 0 {
			return authHeaders[0], nil
		}
	}

	return "", fmt.Errorf("no authorization token found in context")
}

// addTokenToOutgoingContext добавляет токен в исходящие metadata
func addTokenToOutgoingContext(ctx context.Context, token string) context.Context {
	// Получаем существующие outgoing metadata (если есть)
	md, ok := metadata.FromOutgoingContext(ctx)
	if !ok {
		// Если нет - создаём новые
		md = metadata.New(nil)
	} else {
		// Копируем существующие (чтобы не мутировать оригинал)
		md = md.Copy()
	}

	// Добавляем токен
	md.Set("authorization", token)

	// Создаём новый context с обновлёнными metadata
	return metadata.NewOutgoingContext(ctx, md)
}
