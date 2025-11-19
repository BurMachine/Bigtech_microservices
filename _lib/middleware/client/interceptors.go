package platform_client

import (
	"context"
	"math/rand"
	"time"

	"github.com/Burmachine/MSA/lib/metrics" // ← Добавили
	platform_middleware "github.com/Burmachine/MSA/lib/middleware"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_retry "github.com/grpc-ecosystem/go-grpc-middleware/retry"
	"github.com/sony/gobreaker"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
)

// NewClientConn создает gRPC client connection с метриками, retry и circuit breaker
func NewClientConn(addr string, targetService string, m *metrics.Metrics, cfg *platform_middleware.ClientGRPCConfig) (*grpc.ClientConn, error) {
	timeoutDur, _ := time.ParseDuration(cfg.Timeout)
	baseDur, _ := time.ParseDuration(cfg.Retry.Backoff.Base)
	maxDur, _ := time.ParseDuration(cfg.Retry.Backoff.Max)
	windowDur, _ := time.ParseDuration(cfg.CircuitBreaker.Window)
	openDur, _ := time.ParseDuration(cfg.CircuitBreaker.OpenStateFor)

	var unaryClientInterceptors []grpc.UnaryClientInterceptor
	var streamClientInterceptors []grpc.StreamClientInterceptor

	// 1. Tracing interceptor (первым, чтобы создать span)
	unaryClientInterceptors = append(unaryClientInterceptors, NewClientTracingInterceptor())

	// 2. Metrics interceptor (после tracing, до retry/circuit breaker)
	if m != nil {
		unaryClientInterceptors = append(unaryClientInterceptors, NewClientMetricsInterceptor(m, targetService))
		streamClientInterceptors = append(streamClientInterceptors, NewClientMetricsStreamInterceptor(m, targetService))
	}

	// 3. Retry: проверка maxAttempts > 0, codes len > 0, baseDur > 0
	if cfg.Retry.MaxAttempts > 0 && len(cfg.Retry.RetryableCodes) > 0 && baseDur > 0 {
		var retryCodes []codes.Code
		for _, codeStr := range cfg.Retry.RetryableCodes {
			if code, ok := codeMap[codeStr]; ok {
				retryCodes = append(retryCodes, code)
			}
		}

		retryOpts := []grpc_retry.CallOption{
			grpc_retry.WithMax(cfg.Retry.MaxAttempts),
			grpc_retry.WithBackoff(grpc_retry.BackoffExponential(baseDur)),
			grpc_retry.WithCodes(retryCodes...),
		}
		if timeoutDur > 0 { // Timeout per retry, если >0
			retryOpts = append(retryOpts, grpc_retry.WithPerRetryTimeout(timeoutDur))
		}
		if cfg.Retry.Backoff.Jitter {
			retryOpts = append(retryOpts, grpc_retry.WithBackoff(func(attempt uint) time.Duration {
				backoff := baseDur * (1 << attempt)
				if backoff > maxDur {
					backoff = maxDur
				}
				return backoff/2 + time.Duration(rand.Int63n(int64(backoff)))
			}))
		}

		unaryClientInterceptors = append(unaryClientInterceptors, grpc_retry.UnaryClientInterceptor(retryOpts...))
	}

	// 4. Circuit Breaker: проверка failuresForOpen > 0, window > 0, halfOpen > 0
	if cfg.CircuitBreaker.FailuresForOpen > 0 && windowDur > 0 && cfg.CircuitBreaker.HalfOpenMaxCalls > 0 {
		cb := gobreaker.NewCircuitBreaker(gobreaker.Settings{
			Name:        "grpc-client-" + targetService,
			MaxRequests: uint32(cfg.CircuitBreaker.HalfOpenMaxCalls),
			Interval:    windowDur,
			Timeout:     openDur,
			ReadyToTrip: func(counts gobreaker.Counts) bool {
				return counts.ConsecutiveFailures > uint32(cfg.CircuitBreaker.FailuresForOpen)
			},
		})

		cbInterceptor := func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
			_, err := cb.Execute(func() (interface{}, error) {
				return nil, invoker(ctx, method, req, reply, cc, opts...)
			})
			return err
		}

		unaryClientInterceptors = append(unaryClientInterceptors, cbInterceptor)
	}

	// Собираем dial options
	dialOpts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}

	if len(unaryClientInterceptors) > 0 {
		dialOpts = append(dialOpts, grpc.WithUnaryInterceptor(grpc_middleware.ChainUnaryClient(unaryClientInterceptors...)))
	}

	if len(streamClientInterceptors) > 0 {
		dialOpts = append(dialOpts, grpc.WithStreamInterceptor(grpc_middleware.ChainStreamClient(streamClientInterceptors...)))
	}

	return grpc.NewClient(addr, dialOpts...)
}

var codeMap = map[string]codes.Code{
	"OK":                  codes.OK,
	"CANCELLED":           codes.Canceled,
	"UNKNOWN":             codes.Unknown,
	"INVALID_ARGUMENT":    codes.InvalidArgument,
	"DEADLINE_EXCEEDED":   codes.DeadlineExceeded,
	"NOT_FOUND":           codes.NotFound,
	"ALREADY_EXISTS":      codes.AlreadyExists,
	"PERMISSION_DENIED":   codes.PermissionDenied,
	"RESOURCE_EXHAUSTED":  codes.ResourceExhausted,
	"FAILED_PRECONDITION": codes.FailedPrecondition,
	"ABORTED":             codes.Aborted,
	"OUT_OF_RANGE":        codes.OutOfRange,
	"UNIMPLEMENTED":       codes.Unimplemented,
	"INTERNAL":            codes.Internal,
	"UNAVAILABLE":         codes.Unavailable,
	"DATA_LOSS":           codes.DataLoss,
	"UNAUTHENTICATED":     codes.Unauthenticated,
}
