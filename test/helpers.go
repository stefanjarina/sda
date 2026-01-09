package test

import (
	"os"
	"path/filepath"
)

func CreateTempConfig(t interface{ TempDir() string }, content string) string {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "sda.yaml")
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		panic(err)
	}
	return configPath
}

func CreateTempDir(t interface{ TempDir() string }) string {
	tmpDir, err := os.MkdirTemp("", "sda-test-*")
	if err != nil {
		panic(err)
	}
	return tmpDir
}
