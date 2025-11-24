package loggerlib

import (
	"context"
	"fmt"
	"runtime"
	"strings"

	"github.com/opentracing/opentracing-go"
	"github.com/uber/jaeger-client-go"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Level представляет уровень логирования
type Level int

const (
	DebugLevel Level = iota
	InfoLevel
	WarnLevel
	ErrorLevel
	FatalLevel
)

type Logger struct {
	zap         *zap.Logger
	serviceName string
	version     string
	environment string
	level       zapcore.Level // Добавили поле для хранения уровня
}

type Config struct {
	ServiceName string
	Version     string
	Environment string
	Level       string
}

func New(cfg Config) (*Logger, error) {
	zapCfg := zap.NewProductionConfig()
	logLevel := parseLevel(cfg.Level)
	zapCfg.Level = zap.NewAtomicLevelAt(logLevel)
	zapCfg.EncoderConfig.TimeKey = "timestamp"
	zapCfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	zapCfg.EncoderConfig.MessageKey = "message"
	zapCfg.EncoderConfig.LevelKey = "level"
	zapCfg.EncoderConfig.CallerKey = "caller"
	zapCfg.EncoderConfig.EncodeCaller = zapcore.ShortCallerEncoder

	z, err := zapCfg.Build(
		zap.AddCallerSkip(1),
	)
	if err != nil {
		return nil, err
	}

	z = z.With(
		zap.String("service.name", cfg.ServiceName),
		zap.String("service.version", cfg.Version),
		zap.String("env", cfg.Environment),
	)

	return &Logger{
		zap:         z,
		serviceName: cfg.ServiceName,
		version:     cfg.Version,
		environment: cfg.Environment,
		level:       logLevel, // Сохраняем уровень
	}, nil
}

func NewFromZap(zapLogger *zap.Logger, cfg Config) *Logger {
	zapLogger = zapLogger.With(
		zap.String("service.name", cfg.ServiceName),
		zap.String("service.version", cfg.Version),
		zap.String("env", cfg.Environment),
	)

	return &Logger{
		zap:         zapLogger,
		serviceName: cfg.ServiceName,
		version:     cfg.Version,
		environment: cfg.Environment,
		level:       parseLevel(cfg.Level), // Сохраняем уровень
	}
}

// Level возвращает текущий уровень логирования в нашем формате
func (l *Logger) Level() Level {
	switch l.level {
	case zap.DebugLevel:
		return DebugLevel
	case zap.InfoLevel:
		return InfoLevel
	case zap.WarnLevel:
		return WarnLevel
	case zap.ErrorLevel:
		return ErrorLevel
	case zap.FatalLevel:
		return FatalLevel
	default:
		return InfoLevel
	}
}

// ZapLevel возвращает текущий уровень логирования в формате Zap
func (l *Logger) ZapLevel() zapcore.Level {
	return l.level
}

// IsDebugEnabled проверяет, включен ли уровень Debug
func (l *Logger) IsDebugEnabled() bool {
	return l.level <= zap.DebugLevel
}

// IsInfoEnabled проверяет, включен ли уровень Info
func (l *Logger) IsInfoEnabled() bool {
	return l.level <= zap.InfoLevel
}

func (l *Logger) Debug(ctx context.Context, msg string, keysAndValues ...interface{}) {
	fields := l.buildFields(ctx, keysAndValues)
	l.zap.Debug(msg, fields...)
}

func (l *Logger) Info(ctx context.Context, msg string, keysAndValues ...interface{}) {
	fields := l.buildFields(ctx, keysAndValues)
	l.zap.Info(msg, fields...)
}

func (l *Logger) Warn(ctx context.Context, msg string, keysAndValues ...interface{}) {
	fields := l.buildFields(ctx, keysAndValues)
	l.zap.Warn(msg, fields...)
}

func (l *Logger) Error(ctx context.Context, msg string, keysAndValues ...interface{}) {
	fields := l.buildFields(ctx, keysAndValues)
	l.zap.Error(msg, fields...)
}

func (l *Logger) Fatal(ctx context.Context, msg string, keysAndValues ...interface{}) {
	fields := l.buildFields(ctx, keysAndValues)
	l.zap.Fatal(msg, fields...)
}

func (l *Logger) With(keysAndValues ...interface{}) *Logger {
	fields := l.keysAndValuesToZap(keysAndValues)
	return &Logger{
		zap:         l.zap.With(fields...),
		serviceName: l.serviceName,
		version:     l.version,
		environment: l.environment,
		level:       l.level, // Копируем уровень
	}
}

func (l *Logger) WithError(err error) *Logger {
	return &Logger{
		zap:         l.zap.With(zap.Error(err)),
		serviceName: l.serviceName,
		version:     l.version,
		environment: l.environment,
		level:       l.level, // Копируем уровень
	}
}

func (l *Logger) buildFields(ctx context.Context, keysAndValues []interface{}) []zap.Field {
	fields := make([]zap.Field, 0)

	// 1. Добавляем контекстные поля (trace_id, span_id, req_id)
	fields = append(fields, l.contextFields(ctx)...)

	// 2. Добавляем caller (файл:строка)
	if caller := l.getCaller(); caller != "" {
		fields = append(fields, zap.String("caller", caller))
	}

	// 3. Добавляем пользовательские key-value пары
	fields = append(fields, l.keysAndValuesToZap(keysAndValues)...)

	return fields
}

// contextFields извлекает поля из контекста
func (l *Logger) contextFields(ctx context.Context) []zap.Field {
	if ctx == nil {
		return nil
	}

	fields := make([]zap.Field, 0, 3)

	// Извлекаем trace_id и span_id из OpenTracing (Jaeger)
	span := opentracing.SpanFromContext(ctx)
	if span != nil {
		// Приводим к Jaeger SpanContext
		if sc, ok := span.Context().(jaeger.SpanContext); ok {
			fields = append(fields,
				zap.String("trace_id", sc.TraceID().String()),
				zap.String("span_id", sc.SpanID().String()),
			)
		}
	}

	// Извлекаем request_id если есть
	if reqID := getRequestIDFromContext(ctx); reqID != "" {
		fields = append(fields, zap.String("req_id", reqID))
	}

	// Извлекаем user_id если есть
	if userID := getUserIDFromContext(ctx); userID != "" {
		fields = append(fields, zap.String("user_id", userID))
	}

	return fields
}

func (l *Logger) keysAndValuesToZap(keysAndValues []interface{}) []zap.Field {
	if len(keysAndValues) == 0 {
		return nil
	}

	if len(keysAndValues)%2 != 0 {
		keysAndValues = append(keysAndValues, "MISSING_VALUE")
	}

	fields := make([]zap.Field, 0, len(keysAndValues)/2)
	for i := 0; i < len(keysAndValues); i += 2 {
		key, ok := keysAndValues[i].(string)
		if !ok {
			continue
		}

		value := keysAndValues[i+1]

		if err, ok := value.(error); ok && err != nil {
			fields = append(fields, zap.Error(err))
			continue
		}

		fields = append(fields, zap.Any(key, value))
	}

	return fields
}

func (l *Logger) getCaller() string {
	_, file, line, ok := runtime.Caller(4)
	if !ok {
		return ""
	}

	if idx := strings.LastIndexByte(file, '/'); idx != -1 {
		file = file[idx+1:]
	}

	return fmt.Sprintf("%s:%d", file, line)
}

func parseLevel(level string) zapcore.Level {
	switch strings.ToLower(level) {
	case "debug":
		return zap.DebugLevel
	case "info":
		return zap.InfoLevel
	case "warn", "warning":
		return zap.WarnLevel
	case "error":
		return zap.ErrorLevel
	case "fatal":
		return zap.FatalLevel
	default:
		return zap.InfoLevel
	}
}

type contextKey string

const (
	requestIDKey contextKey = "request_id"
	userIDKey    contextKey = "user_id"
	debugKey     contextKey = "debug_enabled"
)

func WithRequestID(ctx context.Context, reqID string) context.Context {
	return context.WithValue(ctx, requestIDKey, reqID)
}

func WithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, userIDKey, userID)
}

func WithDebug(ctx context.Context, enabled bool) context.Context {
	return context.WithValue(ctx, debugKey, enabled)
}

func getRequestIDFromContext(ctx context.Context) string {
	if reqID, ok := ctx.Value(requestIDKey).(string); ok {
		return reqID
	}
	return ""
}

func getUserIDFromContext(ctx context.Context) string {
	if userID, ok := ctx.Value(userIDKey).(string); ok {
		return userID
	}
	return ""
}

func IsDebugFromContext(ctx context.Context) bool {
	if debug, ok := ctx.Value(debugKey).(bool); ok {
		return debug
	}
	return false
}

func (l *Logger) Sync() error {
	return l.zap.Sync()
}
