package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/viper"
)

func TestGetServiceByName_Exists(t *testing.T) {
	CONFIG = Config{
		Network:  "sda-network",
		Password: "password",
		Prefix:   "sda-",
		Services: []Service{
			{
				Name:    "postgres",
				Version: "15",
				Docker: Docker{
					ImageName: "postgres:15",
				},
			},
			{
				Name:    "redis",
				Version: "7",
				Docker: Docker{
					ImageName: "redis:7",
				},
			},
		},
	}

	result := CONFIG.GetServiceByName("postgres")

	if result == nil {
		t.Fatal("Expected service to be found")
	}
	if result.Name != "postgres" {
		t.Errorf("Expected name 'postgres', got '%s'", result.Name)
	}
}

func TestGetServiceByName_NotExists(t *testing.T) {
	CONFIG = Config{
		Services: []Service{
			{Name: "postgres"},
		},
	}

	result := CONFIG.GetServiceByName("nonexistent")

	if result != nil {
		t.Error("Expected nil for nonexistent service")
	}
}

func TestGetAllServiceNames(t *testing.T) {
	CONFIG = Config{
		Services: []Service{
			{Name: "postgres"},
			{Name: "redis"},
			{Name: "mssql"},
		},
	}

	names := CONFIG.GetAllServiceNames()

	if len(names) != 3 {
		t.Errorf("Expected 3 names, got %d", len(names))
	}
	if names[0] != "postgres" || names[1] != "redis" || names[2] != "mssql" {
		t.Error("Names not in expected order")
	}
}

func TestServiceExists(t *testing.T) {
	tests := []struct {
		name     string
		service  string
		expected bool
	}{
		{"existing service", "postgres", true},
		{"nonexistent service", "mongodb", false},
		{"empty string", "", false},
	}

	CONFIG = Config{
		Services: []Service{
			{Name: "postgres"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CONFIG.ServiceExists(tt.service)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestUpdateNetwork(t *testing.T) {
	CONFIG = Config{Network: "old-network"}

	CONFIG.UpdateNetwork("new-network")

	if CONFIG.Network != "new-network" {
		t.Errorf("Expected 'new-network', got '%s'", CONFIG.Network)
	}
}

func TestUpdatePassword(t *testing.T) {
	CONFIG = Config{Password: "old-password"}

	CONFIG.UpdatePassword("new-password")

	if CONFIG.Password != "new-password" {
		t.Errorf("Expected 'new-password', got '%s'", CONFIG.Password)
	}
}

func TestUpdateVersion(t *testing.T) {
	CONFIG = Config{
		Services: []Service{
			{Name: "postgres", Version: "15"},
			{Name: "redis", Version: "6"},
		},
	}

	CONFIG.UpdateVersion("redis", "7")

	redis := CONFIG.GetServiceByName("redis")
	if redis.Version != "7" {
		t.Errorf("Expected version '7', got '%s'", redis.Version)
	}
}

func TestConfigFromViper(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "sda.yaml")
	configContent := `
defaultNetwork: test-network
defaultPassword: testpass
prefix: test-
services:
  - name: mssql
    defaultVersion: "2022"
    hasPassword: true
    docker:
      imageName: mcr.microsoft.com/mssql/server:2022
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatal(err)
	}

	viper.SetConfigFile(configPath)
	if err := viper.ReadInConfig(); err != nil {
		t.Fatal(err)
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		t.Fatal(err)
	}

	if cfg.Network != "test-network" {
		t.Errorf("Expected 'test-network', got '%s'", cfg.Network)
	}
	if len(cfg.Services) != 1 {
		t.Errorf("Expected 1 service, got %d", len(cfg.Services))
	}
}

func TestGetServiceByName_FirstService(t *testing.T) {
	CONFIG = Config{
		Services: []Service{
			{Name: "postgres"},
			{Name: "redis"},
		},
	}

	result := CONFIG.GetServiceByName("postgres")

	if result == nil {
		t.Fatal("Expected first service to be found")
	}
	if result.Name != "postgres" {
		t.Errorf("Expected 'postgres', got '%s'", result.Name)
	}
}

func TestGetServiceByName_LastService(t *testing.T) {
	CONFIG = Config{
		Services: []Service{
			{Name: "postgres"},
			{Name: "redis"},
			{Name: "mssql"},
		},
	}

	result := CONFIG.GetServiceByName("mssql")

	if result == nil {
		t.Fatal("Expected last service to be found")
	}
	if result.Name != "mssql" {
		t.Errorf("Expected 'mssql', got '%s'", result.Name)
	}
}

func TestGetAllServiceNames_Empty(t *testing.T) {
	CONFIG = Config{
		Services: []Service{},
	}

	names := CONFIG.GetAllServiceNames()

	if len(names) != 0 {
		t.Errorf("Expected 0 names, got %d", len(names))
	}
}

func TestServiceExists_Empty(t *testing.T) {
	CONFIG = Config{
		Services: []Service{},
	}

	if CONFIG.ServiceExists("postgres") {
		t.Error("Expected false for empty services")
	}
}

func TestUpdateVersion_NotExists(t *testing.T) {
	originalVersion := "15"
	CONFIG = Config{
		Services: []Service{
			{Name: "postgres", Version: originalVersion},
		},
	}

	CONFIG.UpdateVersion("nonexistent", "99")

	postgres := CONFIG.GetServiceByName("postgres")
	if postgres.Version != originalVersion {
		t.Errorf("Version should not change for nonexistent service")
	}
}
