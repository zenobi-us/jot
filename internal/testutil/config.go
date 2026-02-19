package testutil

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/zenobi-us/jot/internal/services"
)

// CreateTestConfig creates a test config file in a temporary directory.
// Returns the path to the config file.
func CreateTestConfig(t *testing.T, dir string, cfg services.Config) string {
	t.Helper()

	configPath := filepath.Join(dir, "jot", "config.json")

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		t.Fatalf("failed to create config directory: %v", err)
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		t.Fatalf("failed to marshal config: %v", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}

	return configPath
}

// SetupTestEnv sets environment variables for the duration of the test.
// Variables are automatically unset when the test completes.
func SetupTestEnv(t *testing.T, envVars map[string]string) {
	t.Helper()

	for key, value := range envVars {
		t.Setenv(key, value)
	}
}

// CreateInvalidConfig creates an invalid JSON config file for error testing.
func CreateInvalidConfig(t *testing.T, dir string) string {
	t.Helper()

	configPath := filepath.Join(dir, "jot", "config.json")

	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		t.Fatalf("failed to create config directory: %v", err)
	}

	if err := os.WriteFile(configPath, []byte("{ invalid json }"), 0644); err != nil {
		t.Fatalf("failed to write invalid config: %v", err)
	}

	return configPath
}
