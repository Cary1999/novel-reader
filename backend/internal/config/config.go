package config

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	HTTPAddr               string
	DatabaseDSN            string
	JWTSecret              string
	GeneratedJWTSecret     bool
	TokenTTL               time.Duration
	AdminUsername          string
	AdminPassword          string
	GeneratedAdminPassword bool
	UploadDir              string
	MaxUploadBytes         int64
	DBMaxOpenConns         int
	DBMaxIdleConns         int
}

func Load() (Config, error) {
	cfg := Config{
		HTTPAddr:       env("HTTP_ADDR", ":8000"),
		AdminUsername:  env("ADMIN_USERNAME", "admin"),
		UploadDir:      env("UPLOAD_DIR", "data/uploads"),
		MaxUploadBytes: envInt64("MAX_UPLOAD_BYTES", 50*1024*1024),
		DBMaxOpenConns: envInt("DB_MAX_OPEN_CONNS", 10),
		DBMaxIdleConns: envInt("DB_MAX_IDLE_CONNS", 5),
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

	adminPassword := os.Getenv("ADMIN_PASSWORD")
	if adminPassword == "" {
		adminPassword = randomHex(9)
		cfg.GeneratedAdminPassword = true
	}
	cfg.AdminPassword = adminPassword

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
