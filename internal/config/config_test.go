package config

import (
	"os"
	"strings"
	"testing"
)

func TestConfigValidate(t *testing.T) {
	tests := []struct {
		name    string
		cfg     *Config
		wantErr string
	}{
		{
			name:    "empty task",
			cfg:     &Config{APIKey: "key"},
			wantErr: "task is required",
		},
		{
			name:    "empty api key",
			cfg:     &Config{Task: "do something"},
			wantErr: "api-key is required",
		},
		{
			name:    "both empty",
			cfg:     &Config{},
			wantErr: "task is required",
		},
		{
			name:    "both set",
			cfg:     &Config{Task: "open firefox", APIKey: "sk-test"},
			wantErr: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.validate()
			if tt.wantErr == "" {
				if err != nil {
					t.Errorf("validate() = %v, want nil", err)
				}
			} else {
				if err == nil {
					t.Errorf("validate() = nil, want error containing %q", tt.wantErr)
				} else if !strings.Contains(err.Error(), tt.wantErr) {
					t.Errorf("validate() = %v, want error containing %q", err, tt.wantErr)
				}
			}
		})
	}
}

func TestEnvOr(t *testing.T) {
	key := "WAYGENT_TEST_ENV_OR"

	original := os.Getenv(key)
	os.Unsetenv(key)
	defer os.Setenv(key, original)

	got := envOr(key, "fallback")
	if got != "fallback" {
		t.Errorf("envOr unset = %q, want %q", got, "fallback")
	}

	os.Setenv(key, "fromenv")
	got = envOr(key, "fallback")
	if got != "fromenv" {
		t.Errorf("envOr set = %q, want %q", got, "fromenv")
	}

	os.Setenv(key, "")
	got = envOr(key, "fallback")
	if got != "fallback" {
		t.Errorf("envOr empty = %q, want %q", got, "fallback")
	}
}

func TestEnvIntOr(t *testing.T) {
	key := "WAYGENT_TEST_ENV_INT"

	original := os.Getenv(key)
	os.Unsetenv(key)
	defer os.Setenv(key, original)

	got := envIntOr(key, 42)
	if got != 42 {
		t.Errorf("envIntOr unset = %d, want %d", got, 42)
	}

	os.Setenv(key, "100")
	got = envIntOr(key, 42)
	if got != 100 {
		t.Errorf("envIntOr set = %d, want %d", got, 100)
	}

	os.Setenv(key, "notanumber")
	got = envIntOr(key, 42)
	if got != 42 {
		t.Errorf("envIntOr invalid = %d, want %d", got, 42)
	}

	os.Setenv(key, "")
	got = envIntOr(key, 42)
	if got != 42 {
		t.Errorf("envIntOr empty = %d, want %d", got, 42)
	}
}

func TestEnvBoolOr(t *testing.T) {
	key := "WAYGENT_TEST_ENV_BOOL"

	original := os.Getenv(key)
	os.Unsetenv(key)
	defer os.Setenv(key, original)

	got := envBoolOr(key, false)
	if got != false {
		t.Errorf("envBoolOr unset = %v, want false", got)
	}

	got = envBoolOr(key, true)
	if got != true {
		t.Errorf("envBoolOr unset fallback = %v, want true", got)
	}

	tests := []struct {
		val   string
		want  bool
		flaky bool
	}{
		{"true", true, false},
		{"True", true, false},
		{"TRUE", true, false},
		{"1", true, false},
		{"false", false, false},
		{"0", false, false},
		{"", true, true},
	}

	for _, tt := range tests {
		t.Run(tt.val, func(t *testing.T) {
			os.Setenv(key, tt.val)
			fallback := true
			if tt.flaky {
				fallback = true
			}
			got := envBoolOr(key, fallback)
			if tt.flaky {
				if got != fallback {
					t.Errorf("envBoolOr(%q) = %v, want fallback %v", tt.val, got, fallback)
				}
			} else if got != tt.want {
				t.Errorf("envBoolOr(%q) = %v, want %v", tt.val, got, tt.want)
			}
		})
	}
}
