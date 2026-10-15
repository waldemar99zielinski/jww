package configuration

import (
	"os"
	"testing"
)

type testConfig struct {
	Name     string `env:"TEST_NAME" bind:"required"`
	Port     int    `env:"TEST_PORT" bind:"optional" default:"8000"`
	Fallback string `env:"TEST_FALLBACK" bind:"optional" default:"hello"`
	NoTag    string
}

func withEnv(t *testing.T, vars map[string]string, fn func()) {
	t.Helper()
	for k, v := range vars {
		t.Setenv(k, v)
	}
	fn()
}

func TestLoadRequiredAndOptional(t *testing.T) {
	withEnv(t, map[string]string{"TEST_NAME": "jww", "TEST_PORT": "9090"}, func() {
		cfg, err := Load[testConfig]()
		if err != nil {
			t.Fatalf("Load: %v", err)
		}
		if cfg.Name != "jww" {
			t.Errorf("Name = %q, want jww", cfg.Name)
		}
		if cfg.Port != 9090 {
			t.Errorf("Port = %d, want 9090", cfg.Port)
		}
		if cfg.Fallback != "hello" {
			t.Errorf("Fallback = %q, want hello (default)", cfg.Fallback)
		}
	})
}

func TestLoadMissingRequiredFails(t *testing.T) {
	os.Unsetenv("TEST_NAME")
	os.Unsetenv("TEST_PORT")
	os.Unsetenv("TEST_FALLBACK")
	_, err := Load[testConfig]()
	if err == nil {
		t.Fatal("expected error for missing required TEST_NAME, got nil")
	}
}

func TestLoadInvalidIntFails(t *testing.T) {
	withEnv(t, map[string]string{"TEST_NAME": "jww", "TEST_PORT": "not-a-number"}, func() {
		_, err := Load[testConfig]()
		if err == nil {
			t.Fatal("expected error for invalid int, got nil")
		}
	})
}

func TestLoadUntaggedFieldLeftZero(t *testing.T) {
	withEnv(t, map[string]string{"TEST_NAME": "jww"}, func() {
		cfg, err := Load[testConfig]()
		if err != nil {
			t.Fatalf("Load: %v", err)
		}
		if cfg.NoTag != "" {
			t.Errorf("NoTag = %q, want empty (untagged field untouched)", cfg.NoTag)
		}
	})
}

func TestLoadNonStructFails(t *testing.T) {
	if _, err := Load[string](); err == nil {
		t.Fatal("expected error loading into a non-struct type, got nil")
	}
}
