package e2e

import (
	"os"
	"testing"

	"github.com/zenobi-us/jot/internal/services"
)

// TestMain is called before any tests run in this package.
// It initializes the logger to respect LOG_LEVEL environment variable.
func TestMain(m *testing.M) {
	// Initialize logger before running tests
	services.InitLogger()

	// Run tests
	exitCode := m.Run()

	// Exit with test result code
	os.Exit(exitCode)
}
