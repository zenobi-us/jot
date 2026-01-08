package services

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	kjson "github.com/knadh/koanf/parsers/json"
	"github.com/knadh/koanf/providers/confmap"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
	"github.com/rs/zerolog"
)

// NotebookConfigFile is the config filename in notebook directories.
const NotebookConfigFile = ".opennotes.json"

// Config represents the global configuration schema.
type Config struct {
	// Notebooks paths are directories containing .opennotes.json
	Notebooks []string `koanf:"notebooks" json:"notebooks"`
	// NotebookPath is the current notebook path (from env, flag, or stored)
	NotebookPath string `koanf:"notebookpath" json:"notebookpath,omitempty"`
}

// ConfigService manages configuration loading and persistence.
type ConfigService struct {
	k     *koanf.Koanf
	Store Config
	path  string
	log   zerolog.Logger
}

// GlobalConfigFile returns the platform-specific config path.
func GlobalConfigFile() string {
	configDir, err := os.UserConfigDir()
	if err != nil {
		configDir = filepath.Join(os.Getenv("HOME"), ".config")
	}
	return filepath.Join(configDir, "opennotes", "config.json")
}

// NewConfigService creates and initializes the config service.
func NewConfigService() (*ConfigService, error) {
	return NewConfigServiceWithPath(GlobalConfigFile())
}

// NewConfigServiceWithPath creates a config service with a custom config path.
// Useful for testing with isolated config files.
func NewConfigServiceWithPath(configPath string) (*ConfigService, error) {
	log := Log("ConfigService")
	k := koanf.New(".")

	log.Debug().Str("path", configPath).Msg("loading config")

	// 1. Load defaults
	defaultNotebooksDir := filepath.Join(filepath.Dir(configPath), "notebooks")
	defaults := map[string]interface{}{
		"notebooks":    []string{defaultNotebooksDir},
		"notebookpath": "",
	}

	if err := k.Load(confmap.Provider(defaults, "."), nil); err != nil {
		return nil, fmt.Errorf("failed to load defaults: %w", err)
	}

	// 2. Load from config file (if exists)
	if _, err := os.Stat(configPath); err == nil {
		if err := k.Load(file.Provider(configPath), kjson.Parser()); err != nil {
			log.Warn().Err(err).Msg("failed to load config file, using defaults")
		}
	}

	// 3. Load environment variables with OPENNOTES_ prefix
	// Transform: OPENNOTES_NOTEBOOK_PATH -> notebookpath
	err := k.Load(env.Provider("OPENNOTES_", ".", func(s string) string {
		return strings.ToLower(
			strings.ReplaceAll(
				strings.TrimPrefix(s, "OPENNOTES_"),
				"_",
				"",
			),
		)
	}), nil)
	if err != nil {
		log.Warn().Err(err).Msg("failed to load env vars")
	}

	// 4. Unmarshal to struct
	var cfg Config
	if err := k.Unmarshal("", &cfg); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	log.Debug().Interface("config", cfg).Msg("config loaded")

	return &ConfigService{
		k:     k,
		Store: cfg,
		path:  configPath,
		log:   log,
	}, nil
}

// Write persists the configuration to disk.
func (c *ConfigService) Write(cfg Config) error {
	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(c.path), 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(c.path, data, 0644); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	c.Store = cfg
	c.log.Debug().Str("path", c.path).Msg("config written")

	return nil
}

// Path returns the config file path.
func (c *ConfigService) Path() string {
	return c.path
}
