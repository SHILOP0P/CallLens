package logger

import (
	"context"
	"testing"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

func TestLoggerDoesNotDuplicateExplicitUserID(t *testing.T) {
	core, logs := observer.New(zapcore.InfoLevel)
	log := &zapLogger{logger: zap.New(core)}

	ctx := ContextWithUserID(context.Background(), "from-context")
	log.Info(ctx, "message", zap.String("user_id", "explicit"))

	entries := logs.All()
	if len(entries) != 1 {
		t.Fatalf("logged entries = %d, want 1", len(entries))
	}

	userIDFields := 0
	for _, field := range entries[0].Context {
		if field.Key == "user_id" {
			userIDFields++
		}
	}

	if userIDFields != 1 {
		t.Fatalf("user_id fields = %d, want 1", userIDFields)
	}

	if got := entries[0].ContextMap()["user_id"]; got != "explicit" {
		t.Fatalf("user_id = %v, want explicit", got)
	}
}

func TestLoggerLevelsContextsAndMethods(t *testing.T) {
	for _, tt := range []struct {
		input string
		want  zapcore.Level
	}{
		{"debug", zapcore.DebugLevel},
		{"INFO", zapcore.InfoLevel},
		{"warn", zapcore.WarnLevel},
		{"warning", zapcore.WarnLevel},
		{"error", zapcore.ErrorLevel},
		{"unknown", zapcore.InfoLevel},
	} {
		if got := parseLevel(tt.input); got != tt.want {
			t.Fatalf("parseLevel(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}

	ctx := ContextWithTraceID(context.Background(), "trace")
	ctx = ContextWithUserID(ctx, "user")
	fields := fieldsFromContext(ctx)
	if len(fields) != 2 {
		t.Fatalf("context fields = %d", len(fields))
	}
	if got := fieldsWithContext(ctx, nil); len(got) != 2 {
		t.Fatalf("fieldsWithContext nil = %d", len(got))
	}
	if got := fieldsWithContext(ctx, []zap.Field{zap.String("trace_id", "explicit")}); len(got) != 2 {
		t.Fatalf("fieldsWithContext explicit = %d", len(got))
	}

	background := context.Background()
	if ContextWithTraceID(background, "") != background || ContextWithUserID(background, "") != background {
		t.Fatal("empty context values should preserve the original context")
	}

	core, logs := observer.New(zapcore.DebugLevel)
	log := &zapLogger{logger: zap.New(core)}
	log.Debug(ctx, "debug")
	log.Info(ctx, "info")
	log.Warn(ctx, "warn")
	log.Error(ctx, "error")
	if logs.Len() != 4 {
		t.Fatalf("logged entries = %d", logs.Len())
	}

	nop := NewNop()
	nop.Debug(ctx, "debug")
	nop.Info(ctx, "info")
	nop.Warn(ctx, "warn")
	nop.Error(ctx, "error")

	if New("info", false) == nil || New("debug", true) == nil {
		t.Fatal("New returned nil")
	}
	if cfg := buildEncoderConfig(); cfg.MessageKey != "message" {
		t.Fatalf("encoder config = %+v", cfg)
	}
}
