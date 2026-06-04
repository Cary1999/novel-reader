package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadDefaultMaxUploadBytes(t *testing.T) {
	t.Setenv("DATABASE_DSN", "user:pass@tcp(127.0.0.1:3306)/novel_reader")
	t.Setenv("JWT_SECRET", "test-secret")
	t.Setenv("SUPER_ADMIN_PASSWORD", "password123")

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
	t.Setenv("SUPER_ADMIN_PASSWORD", "password123")
	t.Setenv("MAX_UPLOAD_BYTES", "1024")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.MaxUploadBytes != 1024 {
		t.Fatalf("MaxUploadBytes = %d, want 1024", cfg.MaxUploadBytes)
	}
}

func TestLoadResolvesRelativeStorageDirsFromProjectRoot(t *testing.T) {
	root := t.TempDir()
	backendDir := filepath.Join(root, "backend")
	if err := os.WriteFile(filepath.Join(root, "AGENTS.md"), []byte("test"), 0o600); err != nil {
		t.Fatalf("write AGENTS.md: %v", err)
	}
	if err := os.MkdirAll(backendDir, 0o755); err != nil {
		t.Fatalf("mkdir backend dir: %v", err)
	}
	t.Chdir(backendDir)
	t.Setenv("DATABASE_DSN", "user:pass@tcp(127.0.0.1:3306)/novel_reader")
	t.Setenv("JWT_SECRET", "test-secret")
	t.Setenv("SUPER_ADMIN_PASSWORD", "password123")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.UploadDir != filepath.Join(root, "data", "uploads") {
		t.Fatalf("UploadDir = %q, want %q", cfg.UploadDir, filepath.Join(root, "data", "uploads"))
	}
	if cfg.CoverDir != filepath.Join(root, "data", "uploads", "covers") {
		t.Fatalf("CoverDir = %q, want %q", cfg.CoverDir, filepath.Join(root, "data", "uploads", "covers"))
	}
}
