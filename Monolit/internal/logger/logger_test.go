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
