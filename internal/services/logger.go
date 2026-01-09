package services

import (
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// InitLogger initializes the global logger with environment-based configuration.
// Supports DEBUG env var for debug level and LOG_LEVEL for custom levels.
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

	// Pretty console output to stderr
	log.Logger = zerolog.New(zerolog.ConsoleWriter{
		Out: os.Stderr,
	}).With().Timestamp().Logger()
}

// Log returns a child logger with namespace.
func Log(namespace string) zerolog.Logger {
	return log.With().Str("namespace", namespace).Logger()
}
