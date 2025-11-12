package platform_server

import (
	"context"

	loggerlib "github.com/Burmachine/MSA/lib/logger"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/uber/jaeger-client-go"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const (
	traceIDKey = "x-trace-id"
)

func NewServerTracingInterceptors(log *loggerlib.Logger) []grpc.UnaryServerInterceptor {
	return []grpc.UnaryServerInterceptor{
		createTracingInterceptor(log),
	}
}

func createTracingInterceptor(log *loggerlib.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		span := opentracing.SpanFromContext(ctx)

		if span == nil {
			span, ctx = opentracing.StartSpanFromContext(ctx, info.FullMethod)
			defer span.Finish()

			// Извлекаем trace_id и добавляем в metadata
			spanContext, ok := span.Context().(jaeger.SpanContext)
			if ok {
				traceID := spanContext.TraceID().String()

				// Outgoing metadata для других сервисов
				ctx = metadata.NewOutgoingContext(ctx, metadata.Pairs(traceIDKey, traceID))

				// Response header для клиента
				header := metadata.New(map[string]string{traceIDKey: traceID})
				if err := grpc.SendHeader(ctx, header); err != nil {
					log.Warn(ctx, "failed to send trace_id header", "error", err)
				}
			}
		}

		// Добавляем теги
		ext.SpanKindRPCServer.Set(span)
		ext.Component.Set(span, "grpc")
		span.SetTag("rpc.service", info.FullMethod)

		// Выполняем handler
		res, err := handler(ctx, req)

		// Логируем результат
		if err != nil {
			ext.Error.Set(span, true)
			span.LogKV("grpc_error", err.Error())

			code := status.Code(err)
			ext.HTTPStatusCode.Set(span, uint16(code))
			span.SetTag("grpc.status_code", code.String())
		} else {
			span.SetTag("grpc.status_code", "OK")
		}

		return res, err
	}
}

func NewServerTracingStreamInterceptors(log *loggerlib.Logger) []grpc.StreamServerInterceptor {
	return []grpc.StreamServerInterceptor{
		createStreamTracingInterceptor(log),
	}
}

func createStreamTracingInterceptor(log *loggerlib.Logger) grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		ctx := ss.Context()

		span := opentracing.SpanFromContext(ctx)
		if span == nil {
			span, ctx = opentracing.StartSpanFromContext(ctx, info.FullMethod)
			defer span.Finish()

			spanContext, ok := span.Context().(jaeger.SpanContext)
			if ok {
				traceID := spanContext.TraceID().String()
				ctx = metadata.NewOutgoingContext(ctx, metadata.Pairs(traceIDKey, traceID))

				header := metadata.New(map[string]string{traceIDKey: traceID})
				if err := ss.SendHeader(header); err != nil {
					log.Warn(ctx, "failed to send trace_id header", "error", err)
				}
			}
		}

		ext.SpanKindRPCServer.Set(span)
		ext.Component.Set(span, "grpc")
		span.SetTag("rpc.service", info.FullMethod)
		span.SetTag("grpc.stream", true)

		wrappedStream := &tracingServerStream{
			ServerStream: ss,
			ctx:          ctx,
		}

		err := handler(srv, wrappedStream)

		if err != nil {
			ext.Error.Set(span, true)
			span.LogKV("grpc_error", err.Error())
			code := status.Code(err)
			ext.HTTPStatusCode.Set(span, uint16(code))
			span.SetTag("grpc.status_code", code.String())
		} else {
			span.SetTag("grpc.status_code", "OK")
		}

		return err
	}
}

type tracingServerStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (s *tracingServerStream) Context() context.Context {
	return s.ctx
}
