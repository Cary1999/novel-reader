package config

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	HTTPAddr                    string
	DatabaseDSN                 string
	JWTSecret                   string
	GeneratedJWTSecret          bool
	TokenTTL                    time.Duration
	SuperAdminUsername          string
	SuperAdminPassword          string
	GeneratedSuperAdminPassword bool
	UploadDir                   string
	CoverDir                    string
	AvatarDir                   string
	MaxUploadBytes              int64
	MaxCoverBytes               int64
	MaxAvatarBytes              int64
	DefaultCoverFile            string
	DBMaxOpenConns              int
	DBMaxIdleConns              int
}

func Load() (Config, error) {
	cfg := Config{
		HTTPAddr:           env("HTTP_ADDR", ":8000"),
		SuperAdminUsername: env("SUPER_ADMIN_USERNAME", env("ADMIN_USERNAME", "admin")),
		UploadDir:          resolveProjectPath(env("UPLOAD_DIR", "data/uploads")),
		CoverDir:           resolveProjectPath(env("COVER_DIR", "data/uploads/covers")),
		AvatarDir:          resolveProjectPath(env("AVATAR_DIR", "data/uploads/avatars")),
		MaxUploadBytes:     envInt64("MAX_UPLOAD_BYTES", 50*1024*1024),
		MaxCoverBytes:      envInt64("MAX_COVER_BYTES", 10*1024*1024),
		MaxAvatarBytes:     envInt64("MAX_AVATAR_BYTES", 5*1024*1024),
		DefaultCoverFile:   env("DEFAULT_COVER_FILE", "base.jpeg"),
		DBMaxOpenConns:     envInt("DB_MAX_OPEN_CONNS", 10),
		DBMaxIdleConns:     envInt("DB_MAX_IDLE_CONNS", 5),
	}
	cfg.DatabaseDSN = os.Getenv("DATABASE_DSN")
	if cfg.DatabaseDSN == "" {
		return Config{}, fmt.Errorf("DATABASE_DSN is required")
	}

	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = randomHex(32)
		cfg.GeneratedJWTSecret = true
	}
	cfg.JWTSecret = secret

	superAdminPassword := env("SUPER_ADMIN_PASSWORD", os.Getenv("ADMIN_PASSWORD"))
	if superAdminPassword == "" {
		superAdminPassword = randomHex(9)
		cfg.GeneratedSuperAdminPassword = true
	}
	cfg.SuperAdminPassword = superAdminPassword

	ttlHours := envInt("TOKEN_TTL_HOURS", 24*7)
	if ttlHours <= 0 {
		return Config{}, fmt.Errorf("TOKEN_TTL_HOURS must be positive")
	}
	cfg.TokenTTL = time.Duration(ttlHours) * time.Hour
	return cfg, nil
}

func env(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func envInt(key string, fallback int) int {
	raw := os.Getenv(key)
	if raw == "" {
		return fallback
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return fallback
	}
	return value
}

func envInt64(key string, fallback int64) int64 {
	raw := os.Getenv(key)
	if raw == "" {
		return fallback
	}
	value, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return fallback
	}
	return value
}

func randomHex(bytesLen int) string {
	buf := make([]byte, bytesLen)
	if _, err := rand.Read(buf); err != nil {
		panic(err)
	}
	return hex.EncodeToString(buf)
}

func resolveProjectPath(path string) string {
	if strings.TrimSpace(path) == "" {
		return path
	}
	if filepath.IsAbs(path) {
		return filepath.Clean(path)
	}
	if root := detectProjectRoot(); root != "" {
		return filepath.Join(root, filepath.Clean(path))
	}
	return filepath.Clean(path)
}

func detectProjectRoot() string {
	cwd, err := os.Getwd()
	if err != nil {
		return ""
	}
	dir := cwd
	for {
		if _, err := os.Stat(filepath.Join(dir, "AGENTS.md")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return ""
		}
		dir = parent
	}
}
