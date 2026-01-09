package docker

import (
	"regexp"
	"testing"

	"github.com/docker/docker/api/types"
	"github.com/stefanjarina/sda/internal/config"
)

func TestGetNameFromContainerName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"with prefix", "/sda-postgres", "postgres"},
		{"single name", "/sda-mssql", "mssql"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config.CONFIG.Prefix = "sda"
			result := getNameFromContainerName(tt.input)
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestGetVersionFromImageName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"with tag", "postgres:15", "15"},
		{"with digest", "postgres@sha256:abc123", "abc123"},
		{"latest tag", "redis:latest", "latest"},
		{"no tag", "mcr.microsoft.com/mssql/server", "mcr.microsoft.com/mssql/server"},
		{"complex version", "neo4j:5.12", "5.12"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getVersionFromImageName(tt.input)
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestGetPortsFromContainer(t *testing.T) {
	tests := []struct {
		name     string
		ports    []types.Port
		expected []string
	}{
		{
			name:     "single port",
			ports:    []types.Port{{PrivatePort: 5432, PublicPort: 5432}},
			expected: []string{"5432:5432"},
		},
		{
			name: "multiple ports",
			ports: []types.Port{
				{PrivatePort: 5432, PublicPort: 5432},
				{PrivatePort: 5433, PublicPort: 5433},
			},
			expected: []string{"5432:5432", "5433:5433"},
		},
		{
			name:     "no ports",
			ports:    []types.Port{},
			expected: []string{},
		},
		{
			name:     "only private port",
			ports:    []types.Port{{PrivatePort: 5432, PublicPort: 0}},
			expected: []string{"0:5432"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getPortsFromContainer(tt.ports)
			if len(result) != len(tt.expected) {
				t.Errorf("Expected %d ports, got %d", len(tt.expected), len(result))
			}
			for i, port := range result {
				if port != tt.expected[i] {
					t.Errorf("Expected '%s', got '%s'", tt.expected[i], port)
				}
			}
		})
	}
}

func TestGetNamedVolumesForService(t *testing.T) {
	service := &config.Service{
		Name: "postgres",
		Docker: config.Docker{
			Volumes: []config.Volume{
				{Source: "postgres-data", Target: "/data", IsNamed: true},
				{Source: "/host/path", Target: "/container/path", IsNamed: false},
				{Source: "another-volume", Target: "/another", IsNamed: true},
			},
		},
	}

	volumes := GetNamedVolumesForService(service)

	if len(volumes) != 2 {
		t.Errorf("Expected 2 named volumes, got %d", len(volumes))
	}
}

func TestReplacePlaceholder(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		obj      map[string]string
		expected string
	}{
		{
			name:     "simple replacement",
			text:     "{{.NAME}}",
			obj:      map[string]string{"NAME": "postgres"},
			expected: "postgres",
		},
		{
			name:     "password replacement",
			text:     "password={{.PASSWORD}}",
			obj:      map[string]string{"PASSWORD": "secret123"},
			expected: "password=secret123",
		},
		{
			name:     "no placeholders",
			text:     "no placeholders",
			obj:      map[string]string{},
			expected: "no placeholders",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := replacePlaceholder(tt.text, tt.obj)
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestReplacePassword_WithCustomPassword(t *testing.T) {
	service := &config.Service{
		Name:           "postgres",
		CustomPassword: "custom-password",
	}

	result := replacePassword("password={{.PASSWORD}}", service, "default-password")

	if result != "password=custom-password" {
		t.Errorf("Expected custom password, got '%s'", result)
	}
}

func TestReplacePassword_WithDefaultPassword(t *testing.T) {
	service := &config.Service{
		Name:           "postgres",
		CustomPassword: "",
	}

	result := replacePassword("password={{.PASSWORD}}", service, "default-password")

	if result != "password=default-password" {
		t.Errorf("Expected default password, got '%s'", result)
	}
}

func TestParseUlimits(t *testing.T) {
	args := []string{"--ulimit nofile=1024:2048", "--ulimit memlock=65536:65536"}

	ulimits := parseUlimits(args)

	if len(ulimits) != 2 {
		t.Errorf("Expected 2 ulimits, got %d", len(ulimits))
	}
	if ulimits[0].Name != "nofile" {
		t.Errorf("Expected 'nofile', got '%s'", ulimits[0].Name)
	}
	if ulimits[0].Soft != 1024 {
		t.Errorf("Expected soft limit 1024, got %d", ulimits[0].Soft)
	}
}

func TestParseUlimit(t *testing.T) {
	text := "--ulimit nofile=1024:2048"

	ulimit := parseUlimit(text)

	if ulimit == nil {
		t.Fatal("Expected ulimit to be parsed")
	}
	if ulimit.Name != "nofile" {
		t.Errorf("Expected 'nofile', got '%s'", ulimit.Name)
	}
	if ulimit.Soft != 1024 {
		t.Errorf("Expected soft limit 1024, got %d", ulimit.Soft)
	}
	if ulimit.Hard != 2048 {
		t.Errorf("Expected hard limit 2048, got %d", ulimit.Hard)
	}
}

func TestParseUlimit_Invalid(t *testing.T) {
	text := "invalid-ulimit-format"

	ulimit := parseUlimit(text)

	if ulimit != nil {
		t.Error("Expected nil for invalid format")
	}
}

func TestGetMatchByName(t *testing.T) {
	r := regexp.MustCompile(`(?P<name>\w+)=(?P<value>\w+)`)
	match := r.FindStringSubmatch("test=value")
	matchNames := r.SubexpNames()

	result := getMatchByName("name", match, matchNames)

	if result != "test" {
		t.Errorf("Expected 'test', got '%s'", result)
	}
}

func TestGetMatchByName_NotFound(t *testing.T) {
	r := regexp.MustCompile(`(?P<name>\w+)=(?P<value>\w+)`)
	match := r.FindStringSubmatch("test=value")
	matchNames := r.SubexpNames()

	result := getMatchByName("notfound", match, matchNames)

	if result != "" {
		t.Errorf("Expected empty string, got '%s'", result)
	}
}
