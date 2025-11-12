package loggerlib

import (
	"context"
	"fmt"
	"runtime"
	"strings"

	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Logger struct {
	zap         *zap.Logger
	serviceName string
	version     string
	environment string
}

type Config struct {
	ServiceName string
	Version     string
	Environment string
	Level       string // debug, info, warn, error
}

// New создает новый logger с нуля
func New(cfg Config) (*Logger, error) {
	// Настройка zap
	zapCfg := zap.NewProductionConfig()
	zapCfg.Level = zap.NewAtomicLevelAt(parseLevel(cfg.Level))
	zapCfg.EncoderConfig.TimeKey = "timestamp"
	zapCfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	zapCfg.EncoderConfig.MessageKey = "message"
	zapCfg.EncoderConfig.LevelKey = "level"
	zapCfg.EncoderConfig.CallerKey = "caller"
	zapCfg.EncoderConfig.EncodeCaller = zapcore.ShortCallerEncoder

	z, err := zapCfg.Build(
		zap.AddCallerSkip(1), // Пропускаем один уровень вызовов (сам logger)
	)
	if err != nil {
		return nil, err
	}

	// Добавляем базовые поля ко всем логам
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
	}, nil
}

// NewFromZap создает logger из существующего zap.Logger (для rk-boot)
func NewFromZap(zapLogger *zap.Logger, cfg Config) *Logger {
	// Оборачиваем существующий logger, добавляя наши поля
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
	}
}

// Debug логирует debug сообщение
func (l *Logger) Debug(ctx context.Context, msg string, keysAndValues ...interface{}) {
	fields := l.buildFields(ctx, keysAndValues)
	l.zap.Debug(msg, fields...)
}

// Info логирует info сообщение
func (l *Logger) Info(ctx context.Context, msg string, keysAndValues ...interface{}) {
	fields := l.buildFields(ctx, keysAndValues)
	l.zap.Info(msg, fields...)
}

// Warn логирует warning сообщение
func (l *Logger) Warn(ctx context.Context, msg string, keysAndValues ...interface{}) {
	fields := l.buildFields(ctx, keysAndValues)
	l.zap.Warn(msg, fields...)
}

// Error логирует error сообщение
func (l *Logger) Error(ctx context.Context, msg string, keysAndValues ...interface{}) {
	fields := l.buildFields(ctx, keysAndValues)
	l.zap.Error(msg, fields...)
}

// Fatal логирует fatal сообщение и завершает программу
func (l *Logger) Fatal(ctx context.Context, msg string, keysAndValues ...interface{}) {
	fields := l.buildFields(ctx, keysAndValues)
	l.zap.Fatal(msg, fields...)
}

// With создает child logger с дополнительными полями
func (l *Logger) With(keysAndValues ...interface{}) *Logger {
	fields := l.keysAndValuesToZap(keysAndValues)
	return &Logger{
		zap:         l.zap.With(fields...),
		serviceName: l.serviceName,
		version:     l.version,
		environment: l.environment,
	}
}

// WithError создает child logger с полем error
func (l *Logger) WithError(err error) *Logger {
	return &Logger{
		zap:         l.zap.With(zap.Error(err)),
		serviceName: l.serviceName,
		version:     l.version,
		environment: l.environment,
	}
}

// buildFields собирает все поля для лога
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

	// Извлекаем trace_id и span_id из OpenTelemetry
	span := trace.SpanFromContext(ctx)
	if span.SpanContext().IsValid() {
		fields = append(fields,
			zap.String("trace_id", span.SpanContext().TraceID().String()),
			zap.String("span_id", span.SpanContext().SpanID().String()),
		)
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

// keysAndValuesToZap конвертирует key-value пары в zap.Field
func (l *Logger) keysAndValuesToZap(keysAndValues []interface{}) []zap.Field {
	if len(keysAndValues) == 0 {
		return nil
	}

	// Если нечетное количество - добавляем placeholder
	if len(keysAndValues)%2 != 0 {
		keysAndValues = append(keysAndValues, "MISSING_VALUE")
	}

	fields := make([]zap.Field, 0, len(keysAndValues)/2)
	for i := 0; i < len(keysAndValues); i += 2 {
		key, ok := keysAndValues[i].(string)
		if !ok {
			// Если key не string, пропускаем
			continue
		}

		value := keysAndValues[i+1]

		// Специальная обработка для error
		if err, ok := value.(error); ok && err != nil {
			fields = append(fields, zap.Error(err))
			continue
		}

		fields = append(fields, zap.Any(key, value))
	}

	return fields
}

// getCaller возвращает информацию о вызывающей функции
func (l *Logger) getCaller() string {
	// Пропускаем: getCaller -> buildFields -> Info/Error/etc -> user code
	_, file, line, ok := runtime.Caller(4)
	if !ok {
		return ""
	}

	// Получаем только имя файла, без полного пути
	if idx := strings.LastIndexByte(file, '/'); idx != -1 {
		file = file[idx+1:]
	}

	return fmt.Sprintf("%s:%d", file, line)
}

// parseLevel парсит строку уровня в zapcore.Level
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

// Функции для работы с контекстом

type contextKey string

const (
	requestIDKey contextKey = "request_id"
	userIDKey    contextKey = "user_id"
)

// WithRequestID добавляет request_id в контекст
func WithRequestID(ctx context.Context, reqID string) context.Context {
	return context.WithValue(ctx, requestIDKey, reqID)
}

// WithUserID добавляет user_id в контекст
func WithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, userIDKey, userID)
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

// Sync синхронизирует буферы (вызывать перед завершением программы)
func (l *Logger) Sync() error {
	return l.zap.Sync()
}
