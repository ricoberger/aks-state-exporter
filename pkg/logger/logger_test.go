package logger

import (
	"log/slog"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	t.Run("should succeed with valid config", func(t *testing.T) {
		logger := New(Config{Level: slog.LevelDebug, Format: "json"})
		require.NotNil(t, logger)
	})

	t.Run("should succeed with invalid config", func(t *testing.T) {
		logger := New(Config{Level: slog.LevelDebug, Format: "fmt"})
		require.NotNil(t, logger)
	})
}
