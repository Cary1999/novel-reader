package storage

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type SavedFile struct {
	RelativePath string
	Size         int64
	Content      string
}

type LocalStore struct {
	dir      string
	maxBytes int64
}

func NewLocalStore(dir string, maxBytes int64) *LocalStore {
	return &LocalStore{dir: dir, maxBytes: maxBytes}
}

func (s *LocalStore) SaveTXT(originalName string, reader io.Reader) (SavedFile, error) {
	if !strings.EqualFold(filepath.Ext(originalName), ".txt") {
		return SavedFile{}, fmt.Errorf("invalid file type")
	}
	data, err := io.ReadAll(io.LimitReader(reader, s.maxBytes+1))
	if err != nil {
		return SavedFile{}, err
	}
	if int64(len(data)) > s.maxBytes {
		return SavedFile{}, fmt.Errorf("file too large")
	}
	if len(data) == 0 {
		return SavedFile{}, fmt.Errorf("empty file")
	}

	if err := os.MkdirAll(s.dir, 0o755); err != nil {
		return SavedFile{}, err
	}
	filename := time.Now().UTC().Format("20060102T150405Z") + "-" + randomSuffix() + ".txt"
	fullPath := filepath.Join(s.dir, filename)
	if err := os.WriteFile(fullPath, data, 0o600); err != nil {
		return SavedFile{}, err
	}
	return SavedFile{RelativePath: filename, Size: int64(len(data)), Content: string(data)}, nil
}

func randomSuffix() string {
	buf := make([]byte, 6)
	if _, err := rand.Read(buf); err != nil {
		return "fallback"
	}
	return hex.EncodeToString(buf)
}
