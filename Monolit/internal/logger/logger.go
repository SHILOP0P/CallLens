package logger

import (
	"context"
	"os"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Logger interface {
	Debug(ctx context.Context, msg string, fields ...zap.Field)
	Info(ctx context.Context, msg string, fields ...zap.Field)
	Warn(ctx context.Context, msg string, fields ...zap.Field)
	Error(ctx context.Context, msg string, fields ...zap.Field)
}

type contextKey string

const (
	traceIDKey contextKey = "trace_id"
	userIDKey  contextKey = "user_id"
)

type zapLogger struct {
	logger *zap.Logger
}

type nopLogger struct{}

func New(level string, asJSON bool) Logger {
	atomicLevel := zap.NewAtomicLevelAt(parseLevel(level))
	encoderConfig := buildEncoderConfig()

	var encoder zapcore.Encoder
	if asJSON {
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	} else {
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	}

	core := zapcore.NewCore(
		encoder,
		zapcore.AddSync(os.Stdout),
		atomicLevel,
	)

	return &zapLogger{
		logger: zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1)),
	}
}

func NewNop() Logger {
	return nopLogger{}
}

func ContextWithTraceID(ctx context.Context, traceID string) context.Context {
	if traceID == "" {
		return ctx
	}

	return context.WithValue(ctx, traceIDKey, traceID)
}

func ContextWithUserID(ctx context.Context, userID string) context.Context {
	if userID == "" {
		return ctx
	}

	return context.WithValue(ctx, userIDKey, userID)
}

func (l *zapLogger) Debug(ctx context.Context, msg string, fields ...zap.Field) {
	l.logger.Debug(msg, append(fieldsFromContext(ctx), fields...)...)
}

func (l *zapLogger) Info(ctx context.Context, msg string, fields ...zap.Field) {
	l.logger.Info(msg, append(fieldsFromContext(ctx), fields...)...)
}

func (l *zapLogger) Warn(ctx context.Context, msg string, fields ...zap.Field) {
	l.logger.Warn(msg, append(fieldsFromContext(ctx), fields...)...)
}

func (l *zapLogger) Error(ctx context.Context, msg string, fields ...zap.Field) {
	l.logger.Error(msg, append(fieldsFromContext(ctx), fields...)...)
}

func (nopLogger) Debug(context.Context, string, ...zap.Field) {}
func (nopLogger) Info(context.Context, string, ...zap.Field)  {}
func (nopLogger) Warn(context.Context, string, ...zap.Field)  {}
func (nopLogger) Error(context.Context, string, ...zap.Field) {}

func buildEncoderConfig() zapcore.EncoderConfig {
	return zapcore.EncoderConfig{
		TimeKey:        "timestamp",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "message",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
		EncodeName:     zapcore.FullNameEncoder,
	}
}

func parseLevel(level string) zapcore.Level {
	switch strings.ToLower(level) {
	case "debug":
		return zapcore.DebugLevel
	case "info":
		return zapcore.InfoLevel
	case "warn", "warning":
		return zapcore.WarnLevel
	case "error":
		return zapcore.ErrorLevel
	default:
		return zapcore.InfoLevel
	}
}

func fieldsFromContext(ctx context.Context) []zap.Field {
	fields := make([]zap.Field, 0, 2)

	if traceID, ok := ctx.Value(traceIDKey).(string); ok && traceID != "" {
		fields = append(fields, zap.String(string(traceIDKey), traceID))
	}

	if userID, ok := ctx.Value(userIDKey).(string); ok && userID != "" {
		fields = append(fields, zap.String(string(userIDKey), userID))
	}

	return fields
}
