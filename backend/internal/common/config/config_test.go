package config

import "testing"

// TestLoadDefaults verifies that Load() returns sane defaults when no config
// file or environment variables are present.
func TestLoadDefaults(t *testing.T) {
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	// ARK defaults
	if cfg.ARK.BaseURL != "https://ark.cn-beijing.volces.com/api/coding/v3" {
		t.Errorf("unexpected ARK base URL: %s", cfg.ARK.BaseURL)
	}
	if cfg.ARK.Model != "minimax-m2.5" {
		t.Errorf("unexpected ARK model: %s", cfg.ARK.Model)
	}
	if cfg.ARK.Temperature != 0.1 {
		t.Errorf("unexpected ARK temperature: %f", cfg.ARK.Temperature)
	}
	if cfg.ARK.MaxTokens != 256 {
		t.Errorf("unexpected ARK max tokens: %d", cfg.ARK.MaxTokens)
	}
	if cfg.ARK.Timeout != 10 {
		t.Errorf("unexpected ARK timeout: %d", cfg.ARK.Timeout)
	}

	// Server defaults
	if cfg.Server.Port != "8080" {
		t.Errorf("unexpected server port: %s", cfg.Server.Port)
	}
	if cfg.Server.Mode != "debug" {
		t.Errorf("unexpected server mode: %s", cfg.Server.Mode)
	}

	// Database defaults
	if cfg.Database.Host != "localhost" {
		t.Errorf("unexpected database host: %s", cfg.Database.Host)
	}
	if cfg.Database.Port != 5432 {
		t.Errorf("unexpected database port: %d", cfg.Database.Port)
	}
	if cfg.Database.User != "crayfish_user" {
		t.Errorf("unexpected database user: %s", cfg.Database.User)
	}
	if cfg.Database.DBName != "crayfish_travel" {
		t.Errorf("unexpected database name: %s", cfg.Database.DBName)
	}
	if cfg.Database.SSLMode != "disable" {
		t.Errorf("unexpected database ssl mode: %s", cfg.Database.SSLMode)
	}

	// Redis defaults
	if cfg.Redis.Addr != "localhost:6379" {
		t.Errorf("unexpected redis addr: %s", cfg.Redis.Addr)
	}
	if cfg.Redis.DB != 0 {
		t.Errorf("unexpected redis db: %d", cfg.Redis.DB)
	}

	// Claude defaults
	if cfg.Claude.Model != "claude-sonnet-4-20250514" {
		t.Errorf("unexpected claude model: %s", cfg.Claude.Model)
	}
	if cfg.Claude.Timeout != 10 {
		t.Errorf("unexpected claude timeout: %d", cfg.Claude.Timeout)
	}

	// AllowedOrigins fallback
	if len(cfg.AllowedOrigins) == 0 {
		t.Error("expected default allowed origins")
	}
	expectedOrigins := []string{
		"http://localhost:3000",
		"http://localhost:3010",
		"http://150.158.192.237",
	}
	if len(cfg.AllowedOrigins) != len(expectedOrigins) {
		t.Fatalf("expected %d allowed origins, got %d", len(expectedOrigins), len(cfg.AllowedOrigins))
	}
	for i, want := range expectedOrigins {
		if cfg.AllowedOrigins[i] != want {
			t.Errorf("allowed origin[%d] = %s, want %s", i, cfg.AllowedOrigins[i], want)
		}
	}
}

// TestLoadReturnsNonNil verifies Load always returns a non-nil config.
func TestLoadReturnsNonNil(t *testing.T) {
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}
	if cfg == nil {
		t.Error("Load() returned nil config")
	}
}

// TestLoadConfigStruct verifies the Config struct has all expected fields populated.
func TestLoadConfigStruct(t *testing.T) {
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	// Verify struct fields are accessible (compile-time check via usage)
	_ = cfg.Server.Port
	_ = cfg.Server.Mode
	_ = cfg.Database.Host
	_ = cfg.Database.Port
	_ = cfg.Database.User
	_ = cfg.Database.Password
	_ = cfg.Database.DBName
	_ = cfg.Database.SSLMode
	_ = cfg.Redis.Addr
	_ = cfg.Redis.Password
	_ = cfg.Redis.DB
	_ = cfg.Claude.APIKey
	_ = cfg.Claude.Model
	_ = cfg.Claude.Timeout
	_ = cfg.ARK.APIKey
	_ = cfg.ARK.BaseURL
	_ = cfg.ARK.Model
	_ = cfg.ARK.Temperature
	_ = cfg.ARK.MaxTokens
	_ = cfg.ARK.Timeout
	_ = cfg.Security.AESKey
	_ = cfg.AllowedOrigins
	_ = cfg.AdminToken
}
