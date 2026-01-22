package services

import (
	"io"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// InitLogger initializes the global logger with environment-based configuration.
// Supports DEBUG env var for debug level, LOG_LEVEL for custom levels, and LOG_FORMAT for output format.
//
// LOG_FORMAT options:
//   - "compact" (default): Clean, compact console output with short timestamps
//   - "console": Standard colorized console output
//   - "json": Structured JSON output for log aggregation
//   - "ci": Non-colorized console output for CI/CD environments
func InitLogger() {
	// Default to info level
	level := zerolog.InfoLevel

	// Check DEBUG env var
	if os.Getenv("DEBUG") != "" {
		level = zerolog.DebugLevel
	}

	// Check LOG_LEVEL env var (overrides DEBUG)
	if lvl := os.Getenv("LOG_LEVEL"); lvl != "" {
		if parsed, err := zerolog.ParseLevel(lvl); err == nil {
			level = parsed
		}
	}

	zerolog.SetGlobalLevel(level)

	// Configure writer based on LOG_FORMAT
	var writer io.Writer
	format := os.Getenv("LOG_FORMAT")

	switch format {
	case "json":
		// JSON format for structured logging / log aggregation
		writer = os.Stderr

	case "console":
		// Standard console format (original default)
		writer = zerolog.ConsoleWriter{
			Out: os.Stderr,
		}

	case "ci":
		// CI-friendly: no colors, full timestamps
		writer = zerolog.ConsoleWriter{
			Out:        os.Stderr,
			NoColor:    true,
			TimeFormat: time.RFC3339,
		}

	case "compact", "":
		// Compact format (new default): short time, clean output
		writer = zerolog.ConsoleWriter{
			Out:        os.Stderr,
			TimeFormat: "15:04:05",
			PartsOrder: []string{
				zerolog.TimestampFieldName,
				zerolog.LevelFieldName,
				zerolog.MessageFieldName,
			},
		}

	default:
		// Unknown format, fall back to compact
		writer = zerolog.ConsoleWriter{
			Out:        os.Stderr,
			TimeFormat: "15:04:05",
		}
	}

	log.Logger = zerolog.New(writer).With().Timestamp().Logger()
}

// Log returns a child logger with namespace.
func Log(namespace string) zerolog.Logger {
	return log.With().Str("namespace", namespace).Logger()
}
