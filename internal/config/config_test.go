package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/viper"
)

func TestGetServiceByName_Exists(t *testing.T) {
	cfg := Config{
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

	result := cfg.GetServiceByName("postgres")

	if result == nil {
		t.Fatal("Expected service to be found")
	}
	if result.Name != "postgres" {
		t.Errorf("Expected name 'postgres', got '%s'", result.Name)
	}
}

func TestGetServiceByName_ReturnsPointerIntoSlice(t *testing.T) {
	cfg := Config{
		Services: []Service{
			{Name: "postgres", Docker: Docker{}},
			{Name: "redis", Docker: Docker{}},
		},
	}

	result := cfg.GetServiceByName("redis")
	result.Docker.EnvVars = []string{"FOO=bar"}

	again := cfg.GetServiceByName("redis")
	if len(again.Docker.EnvVars) != 1 || again.Docker.EnvVars[0] != "FOO=bar" {
		t.Errorf("Expected mutation through returned pointer to be visible in cfg.Services, got %v", again.Docker.EnvVars)
	}
}

func TestGetServiceByName_NotExists(t *testing.T) {
	cfg := Config{
		Services: []Service{
			{Name: "postgres"},
		},
	}

	result := cfg.GetServiceByName("nonexistent")

	if result != nil {
		t.Error("Expected nil for nonexistent service")
	}
}

func TestGetAllServiceNames(t *testing.T) {
	cfg := Config{
		Services: []Service{
			{Name: "postgres"},
			{Name: "redis"},
			{Name: "mssql"},
		},
	}

	names := cfg.GetAllServiceNames()

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

	cfg := Config{
		Services: []Service{
			{Name: "postgres"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := cfg.ServiceExists(tt.service)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
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
	cfg := Config{
		Services: []Service{
			{Name: "postgres"},
			{Name: "redis"},
		},
	}

	result := cfg.GetServiceByName("postgres")

	if result == nil {
		t.Fatal("Expected first service to be found")
	}
	if result.Name != "postgres" {
		t.Errorf("Expected 'postgres', got '%s'", result.Name)
	}
}

func TestGetServiceByName_LastService(t *testing.T) {
	cfg := Config{
		Services: []Service{
			{Name: "postgres"},
			{Name: "redis"},
			{Name: "mssql"},
		},
	}

	result := cfg.GetServiceByName("mssql")

	if result == nil {
		t.Fatal("Expected last service to be found")
	}
	if result.Name != "mssql" {
		t.Errorf("Expected 'mssql', got '%s'", result.Name)
	}
}

func TestGetAllServiceNames_Empty(t *testing.T) {
	cfg := Config{
		Services: []Service{},
	}

	names := cfg.GetAllServiceNames()

	if len(names) != 0 {
		t.Errorf("Expected 0 names, got %d", len(names))
	}
}

func TestServiceExists_Empty(t *testing.T) {
	cfg := Config{
		Services: []Service{},
	}

	if cfg.ServiceExists("postgres") {
		t.Error("Expected false for empty services")
	}
}
