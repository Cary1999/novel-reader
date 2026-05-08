package config

import "testing"

func TestLoadDefaultMaxUploadBytes(t *testing.T) {
	t.Setenv("DATABASE_DSN", "user:pass@tcp(127.0.0.1:3306)/novel_reader")
	t.Setenv("JWT_SECRET", "test-secret")
	t.Setenv("ADMIN_PASSWORD", "password123")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.MaxUploadBytes != 50*1024*1024 {
		t.Fatalf("MaxUploadBytes = %d, want %d", cfg.MaxUploadBytes, 50*1024*1024)
	}
}

func TestLoadMaxUploadBytesOverride(t *testing.T) {
	t.Setenv("DATABASE_DSN", "user:pass@tcp(127.0.0.1:3306)/novel_reader")
	t.Setenv("JWT_SECRET", "test-secret")
	t.Setenv("ADMIN_PASSWORD", "password123")
	t.Setenv("MAX_UPLOAD_BYTES", "1024")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.MaxUploadBytes != 1024 {
		t.Fatalf("MaxUploadBytes = %d, want 1024", cfg.MaxUploadBytes)
	}
}
