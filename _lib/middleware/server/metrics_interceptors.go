package platform_server

import (
	"context"
	"time"

	"github.com/Burmachine/MSA/lib/metrics"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// NewServerMetricsInterceptors создает unary interceptor для gRPC server метрик
func NewServerMetricsInterceptors(m *metrics.Metrics, serviceName string) []grpc.UnaryServerInterceptor {
	return []grpc.UnaryServerInterceptor{
		func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
			method := info.FullMethod

			// 1. Увеличиваем счетчик активных запросов
			m.GRPCRequestsInFlight.WithLabelValues(serviceName, method).Inc()
			defer m.GRPCRequestsInFlight.WithLabelValues(serviceName, method).Dec()

			// 2. Засекаем время начала
			start := time.Now()

			// 3. Вызываем обработчик
			resp, err := handler(ctx, req)

			// 4. Вычисляем длительность
			duration := time.Since(start)

			// 5. Определяем gRPC status code
			code := codes.OK
			if err != nil {
				st, ok := status.FromError(err)
				if ok {
					code = st.Code()
				} else {
					code = codes.Unknown
				}
			}

			// 6. Записываем метрики
			m.RecordGRPCRequest(serviceName, method, code.String(), duration)

			return resp, err
		},
	}
}

// NewServerMetricsStreamInterceptors создает stream interceptor для gRPC server метрик
func NewServerMetricsStreamInterceptors(m *metrics.Metrics, serviceName string) []grpc.StreamServerInterceptor {
	return []grpc.StreamServerInterceptor{
		func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
			method := info.FullMethod

			// Увеличиваем счетчик активных stream запросов
			m.GRPCRequestsInFlight.WithLabelValues(serviceName, method).Inc()
			defer m.GRPCRequestsInFlight.WithLabelValues(serviceName, method).Dec()

			start := time.Now()
			err := handler(srv, ss)
			duration := time.Since(start)

			code := codes.OK
			if err != nil {
				st, ok := status.FromError(err)
				if ok {
					code = st.Code()
				} else {
					code = codes.Unknown
				}
			}

			m.RecordGRPCRequest(serviceName, method, code.String(), duration)

			return err
		},
	}
}
