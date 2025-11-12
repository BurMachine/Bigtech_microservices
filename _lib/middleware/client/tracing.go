package platform_client

import (
	"context"

	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// NewClientTracingInterceptor создает unary interceptor для трейсинга исходящих gRPC запросов
func NewClientTracingInterceptor() grpc.UnaryClientInterceptor {
	return func(
		ctx context.Context,
		method string,
		req, reply interface{},
		cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker,
		opts ...grpc.CallOption,
	) error {
		// Создаем child span
		span, ctx := opentracing.StartSpanFromContext(ctx, method)
		defer span.Finish()

		// Теги
		ext.SpanKindRPCClient.Set(span)
		ext.Component.Set(span, "grpc-client")
		ext.PeerService.Set(span, cc.Target())
		span.SetTag("rpc.method", method)

		// Inject trace context в gRPC metadata
		md, ok := metadata.FromOutgoingContext(ctx)
		if !ok {
			md = metadata.New(nil)
		}

		carrier := metadataCarrier(md)
		if err := opentracing.GlobalTracer().Inject(
			span.Context(),
			opentracing.TextMap,
			carrier,
		); err == nil {
			ctx = metadata.NewOutgoingContext(ctx, md)
		}

		// Выполняем запрос
		err := invoker(ctx, method, req, reply, cc, opts...)

		// Логируем результат
		if err != nil {
			ext.Error.Set(span, true)

			grpcStatus := status.Convert(err)
			ext.HTTPStatusCode.Set(span, uint16(grpcStatus.Code()))
			span.SetTag("grpc.status_code", grpcStatus.Code().String())
			span.LogKV("grpc_error", grpcStatus.Message())
		} else {
			span.SetTag("grpc.status_code", "OK")
		}

		return err
	}
}

// metadataCarrier адаптер для metadata
type metadataCarrier metadata.MD

func (m metadataCarrier) Set(key, val string) {
	metadata.MD(m).Set(key, val)
}

func (m metadataCarrier) ForeachKey(handler func(key, val string) error) error {
	for k, vals := range metadata.MD(m) {
		for _, v := range vals {
			if err := handler(k, v); err != nil {
				return err
			}
		}
	}
	return nil
}
