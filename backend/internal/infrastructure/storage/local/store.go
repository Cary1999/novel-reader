package local

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	sitecommand "novel-reader/backend/internal/application/site/command"
	uploadapp "novel-reader/backend/internal/application/upload"
)

type Store struct {
	dir      string
	maxBytes int64
}

func NewStore(dir string, maxBytes int64) *Store {
	return &Store{dir: dir, maxBytes: maxBytes}
}

func (s *Store) SaveTXT(originalName string, reader io.Reader) (uploadapp.SavedFile, error) {
	if !strings.EqualFold(filepath.Ext(originalName), ".txt") {
		return uploadapp.SavedFile{}, fmt.Errorf("invalid file type")
	}
	data, err := io.ReadAll(io.LimitReader(reader, s.maxBytes+1))
	if err != nil {
		return uploadapp.SavedFile{}, err
	}
	if int64(len(data)) > s.maxBytes {
		return uploadapp.SavedFile{}, fmt.Errorf("file too large")
	}
	if len(data) == 0 {
		return uploadapp.SavedFile{}, fmt.Errorf("empty file")
	}

	if err := os.MkdirAll(s.dir, 0o755); err != nil {
		return uploadapp.SavedFile{}, err
	}
	filename := time.Now().UTC().Format("20060102T150405Z") + "-" + randomSuffix() + ".txt"
	fullPath := filepath.Join(s.dir, filename)
	if err := os.WriteFile(fullPath, data, 0o600); err != nil {
		return uploadapp.SavedFile{}, err
	}
	return uploadapp.SavedFile{RelativePath: filename, Size: int64(len(data)), Content: string(data)}, nil
}

func (s *Store) SaveCover(originalName string, reader io.Reader) (uploadapp.SavedFile, error) {
	ext := strings.ToLower(filepath.Ext(originalName))
	switch ext {
	case ".jpg", ".jpeg", ".png", ".webp":
	default:
		return uploadapp.SavedFile{}, fmt.Errorf("invalid file type")
	}

	data, err := io.ReadAll(io.LimitReader(reader, s.maxBytes+1))
	if err != nil {
		return uploadapp.SavedFile{}, err
	}
	if int64(len(data)) > s.maxBytes {
		return uploadapp.SavedFile{}, fmt.Errorf("file too large")
	}
	if len(data) == 0 {
		return uploadapp.SavedFile{}, fmt.Errorf("empty file")
	}

	if err := os.MkdirAll(s.dir, 0o755); err != nil {
		return uploadapp.SavedFile{}, err
	}
	filename := time.Now().UTC().Format("20060102T150405Z") + "-" + randomSuffix() + ext
	fullPath := filepath.Join(s.dir, filename)
	if err := os.WriteFile(fullPath, data, 0o600); err != nil {
		return uploadapp.SavedFile{}, err
	}
	return uploadapp.SavedFile{RelativePath: filename, Size: int64(len(data))}, nil
}

func (s *Store) SaveSiteIcon(originalName string, reader io.Reader) (sitecommand.SavedIcon, error) {
	ext := strings.ToLower(filepath.Ext(originalName))
	switch ext {
	case ".jpg", ".jpeg", ".png", ".webp", ".svg":
	default:
		return sitecommand.SavedIcon{}, fmt.Errorf("invalid file type")
	}

	data, err := io.ReadAll(io.LimitReader(reader, s.maxBytes+1))
	if err != nil {
		return sitecommand.SavedIcon{}, err
	}
	if int64(len(data)) > s.maxBytes {
		return sitecommand.SavedIcon{}, fmt.Errorf("file too large")
	}
	if len(data) == 0 {
		return sitecommand.SavedIcon{}, fmt.Errorf("empty file")
	}

	if err := os.MkdirAll(s.dir, 0o755); err != nil {
		return sitecommand.SavedIcon{}, err
	}
	filename := time.Now().UTC().Format("20060102T150405Z") + "-" + randomSuffix() + ext
	fullPath := filepath.Join(s.dir, filename)
	if err := os.WriteFile(fullPath, data, 0o600); err != nil {
		return sitecommand.SavedIcon{}, err
	}
	return sitecommand.SavedIcon{RelativePath: filename, Size: int64(len(data))}, nil
}

func randomSuffix() string {
	buf := make([]byte, 6)
	if _, err := rand.Read(buf); err != nil {
		return "fallback"
	}
	return hex.EncodeToString(buf)
}
