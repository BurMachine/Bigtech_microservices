package platform_client

import (
	"context"
	"time"

	"github.com/Burmachine/MSA/lib/metrics"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// NewClientMetricsInterceptor создает unary interceptor для gRPC client метрик
func NewClientMetricsInterceptor(m *metrics.Metrics, targetService string) grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		// Засекаем время начала
		start := time.Now()

		// Вызываем удаленный метод
		err := invoker(ctx, method, req, reply, cc, opts...)

		// Вычисляем длительность
		duration := time.Since(start)

		// Определяем gRPC status code
		code := codes.OK
		if err != nil {
			st, ok := status.FromError(err)
			if ok {
				code = st.Code()
			} else {
				code = codes.Unknown
			}
		}

		// Записываем метрики
		m.RecordGRPCClientRequest(targetService, method, code.String(), duration)

		return err
	}
}

// NewClientMetricsStreamInterceptor создает stream interceptor для gRPC client метрик
func NewClientMetricsStreamInterceptor(m *metrics.Metrics, targetService string) grpc.StreamClientInterceptor {
	return func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
		start := time.Now()

		clientStream, err := streamer(ctx, desc, cc, method, opts...)

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

		m.RecordGRPCClientRequest(targetService, method, code.String(), duration)

		return clientStream, err
	}
}
