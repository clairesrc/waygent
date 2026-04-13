package config

import (
	"flag"
	"fmt"
	"os"
	"strconv"
)

// Config holds all runtime configuration.
type Config struct {
	Task     string // The task description for the agent
	APIURL   string // OpenAI-compatible API base URL (e.g. "https://api.openai.com/v1")
	APIKey   string // API key for authentication
	Model    string // Model name (e.g. "gpt-4o")
	MaxSteps int    // Maximum agent loop iterations (default 50)
	TempDir  string // Directory for temporary screenshots (default "/tmp/waygent")
	Verbose  bool   // Enable verbose logging
}

func envOr(envKey, fallback string) string {
	if v := os.Getenv(envKey); v != "" {
		return v
	}
	return fallback
}

func envIntOr(envKey string, fallback int) int {
	if v := os.Getenv(envKey); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return fallback
}

func envBoolOr(envKey string, fallback bool) bool {
	if v := os.Getenv(envKey); v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			return b
		}
	}
	return fallback
}

// Load builds Config from CLI flags with env var fallbacks.
func Load() (*Config, error) {
	fs := flag.NewFlagSet("waygent", flag.ContinueOnError)

	task := fs.String("task", envOr("WAYGENT_TASK", ""), "Task description for the agent (required)")
	apiURL := fs.String("api-url", envOr("WAYGENT_API_URL", "https://api.openai.com/v1"), "OpenAI-compatible API base URL")
	apiKey := fs.String("api-key", envOr("WAYGENT_API_KEY", os.Getenv("OPENAI_API_KEY")), "API key for authentication")
	model := fs.String("model", envOr("WAYGENT_MODEL", "gpt-4o"), "Model name")
	maxSteps := fs.Int("max-steps", envIntOr("WAYGENT_MAX_STEPS", 50), "Maximum agent loop iterations")
	verbose := fs.Bool("verbose", envBoolOr("WAYGENT_VERBOSE", false), "Enable verbose logging")

	if err := fs.Parse(os.Args[1:]); err != nil {
		return nil, fmt.Errorf("parsing flags: %w", err)
	}

	cfg := &Config{
		Task:     *task,
		APIURL:   *apiURL,
		APIKey:   *apiKey,
		Model:    *model,
		MaxSteps: *maxSteps,
		TempDir:  envOr("WAYGENT_TEMP_DIR", "/tmp/waygent"),
		Verbose:  *verbose,
	}

	if err := cfg.validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

func (c *Config) validate() error {
	if c.Task == "" {
		return fmt.Errorf("task is required: use -task flag or WAYGENT_TASK env var")
	}
	if c.APIKey == "" {
		return fmt.Errorf("api-key is required: use -api-key flag, WAYGENT_API_KEY, or OPENAI_API_KEY env var")
	}
	return nil
}
