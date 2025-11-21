package dispatcher

import (
	"context"
	"testing"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

func TestLoggerDispatcherDispatch(t *testing.T) {
	core, logs := observer.New(zapcore.InfoLevel)
	logger := zap.New(core)
	d := NewLoggerDispatcher(logger)

	err := d.Dispatch(context.Background(), "user@example.com", "hello")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if logs.Len() != 1 {
		t.Fatalf("expected one log entry, got %d", logs.Len())
	}
	entry := logs.All()[0]
	if entry.ContextMap()["target"] != "user@example.com" || entry.ContextMap()["message"] != "hello" {
		t.Fatalf("log fields missing or wrong")
	}
}
