package middleware_grpc

import (
	"context"
	"errors"
	"log"

	"github.com/BurMachine/Bigtech_microservices/users/internal/app/usecases/users"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ErrorUnaryServerInterceptor - интерцептор для обычных (unary) запросов
func ErrorUnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		resp, err = handler(ctx, req)
		if err != nil {
			return resp, mapError(err, info.FullMethod)
		}
		return resp, nil
	}
}

// ErrorStreamServerInterceptor - интерцептор для потоковых (stream) запросов
func ErrorStreamServerInterceptor() grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		err := handler(srv, ss)
		if err != nil {
			return mapError(err, info.FullMethod)
		}
		return nil
	}
}

// mapError - общая функция маппинга ошибок
func mapError(err error, method string) error {
	// Если ошибка уже является gRPC статусом, возвращаем как есть
	if _, ok := status.FromError(err); ok {
		return err
	}

	// Маппинг бизнес-ошибок на gRPC коды
	switch {
	case errors.Is(err, users.ErrNotFound):
		return status.Error(codes.NotFound, err.Error())

	case errors.Is(err, users.ErrAlreadyExists):
		return status.Error(codes.AlreadyExists, err.Error())

	case errors.Is(err, users.ErrPermissionDenied):
		return status.Error(codes.PermissionDenied, err.Error())

	case errors.Is(err, users.ErrInvalidArgument):
		return status.Error(codes.InvalidArgument, err.Error())

	default:
		// Логируем внутреннюю ошибку с контекстом
		log.Printf("[ERROR] method=%s error=%v", method, err)
		// Не возвращаем детали ошибки клиенту для безопасности
		return status.Error(codes.Internal, "internal server error")
	}
}

// RecoveryUnaryServerInterceptor - интерцептор для обработки паник в unary запросах
func RecoveryUnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("[PANIC] method=%s panic=%v", info.FullMethod, r)
				err = status.Error(codes.Internal, "internal server error")
			}
		}()
		return handler(ctx, req)
	}
}

// RecoveryStreamServerInterceptor - интерцептор для обработки паник в stream запросах
func RecoveryStreamServerInterceptor() grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) (err error) {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("[PANIC] method=%s panic=%v", info.FullMethod, r)
				err = status.Error(codes.Internal, "internal server error")
			}
		}()
		return handler(srv, ss)
	}
}
