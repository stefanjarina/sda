package docker

import (
	"reflect"
	"strings"
	"testing"

	"github.com/stefanjarina/sda/internal/config"
)

func TestGetNameFromContainerName(t *testing.T) {
	d := &Api{cfg: &config.Config{Prefix: "sda"}}
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"with prefix", "sda-postgres", "postgres"},
		{"single name", "sda-mssql", "mssql"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := d.getNameFromContainerName(tt.input)
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
		{"no tag", "mcr.microsoft.com/mssql/server", "latest"},
		{"untagged short name", "postgres", "latest"},
		{"registry with port, no tag", "localhost:5000/postgres", "latest"},
		{"registry with port and tag", "registry.io:5000/team/pg:16", "16"},
		{"ghcr tagged", "ghcr.io/org/app:v1.2.3", "v1.2.3"},
		{"library path untagged", "library/postgres", "latest"},
		{"empty tag", "postgres:", "latest"},
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

func TestParsePorts(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "single port",
			input:    "0.0.0.0:5432->5432/tcp",
			expected: []string{"5432:5432"},
		},
		{
			name:     "ipv4 and ipv6 rows deduplicated",
			input:    "0.0.0.0:5432->5432/tcp, [::]:5432->5432/tcp",
			expected: []string{"5432:5432"},
		},
		{
			name:     "multiple ports",
			input:    "0.0.0.0:5432->5432/tcp, 0.0.0.0:5433->5433/tcp",
			expected: []string{"5432:5432", "5433:5433"},
		},
		{
			name:     "different host port",
			input:    "0.0.0.0:15432->5432/tcp",
			expected: []string{"15432:5432"},
		},
		{
			name:     "no ports",
			input:    "",
			expected: nil,
		},
		{
			name:     "exposed but not published",
			input:    "5432/tcp",
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parsePorts(tt.input)
			if len(result) != len(tt.expected) {
				t.Fatalf("Expected %d ports, got %d (%v)", len(tt.expected), len(result), result)
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

	d := &Api{cfg: &config.Config{Prefix: "sda"}}
	volumes, err := d.GetNamedVolumesForService(service)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := []string{"postgres-data", "another-volume"}
	if !reflect.DeepEqual(volumes, want) {
		t.Fatalf("got %#v, want %#v", volumes, want)
	}
}

func TestGetNamedVolumesForService_CreateAndRemoveAgree(t *testing.T) {
	d := &Api{cfg: &config.Config{Prefix: "sda"}}
	service := &config.Service{
		Name: "postgres",
		Docker: config.Docker{
			Volumes: []config.Volume{
				{Source: "{{.NAME}}-data", Target: "/var/lib/postgresql/data", IsNamed: true},
				{Source: "pgdata", Target: "/data", IsNamed: true},
				{Source: "/host/path", Target: "/container/path", IsNamed: false},
			},
		},
	}

	volumes, err := d.GetNamedVolumesForService(service)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := []string{"sda-postgres-data", "pgdata"}
	if !reflect.DeepEqual(volumes, want) {
		t.Fatalf("GetNamedVolumesForService() = %#v, want %#v", volumes, want)
	}
}

func TestGetNamedVolumesForService_NilService(t *testing.T) {
	d := &Api{cfg: &config.Config{Prefix: "sda"}}
	volumes, err := d.GetNamedVolumesForService(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if volumes != nil {
		t.Fatalf("expected nil, got %v", volumes)
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
			result, err := replacePlaceholder(tt.text, tt.obj)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestReplacePlaceholder_Errors(t *testing.T) {
	tests := []struct {
		name string
		text string
		obj  map[string]string
	}{
		{"unclosed action", "PASS={{.PASSWORD", map[string]string{"PASSWORD": "x"}},
		{"missing key", "PASS={{.PASSWORD}}", map[string]string{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := replacePlaceholder(tt.text, tt.obj)
			if err == nil {
				t.Fatal("expected error")
			}
		})
	}
}

func TestValidateServiceTemplates_BadEnv(t *testing.T) {
	svc := &config.Service{
		Name: "mssql",
		Docker: config.Docker{
			EnvVars: []string{"OK=1", "PASS={{.PASSWORD"},
		},
	}
	err := ValidateServiceTemplates(svc)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "mssql") || !strings.Contains(err.Error(), "envVars[1]") {
		t.Fatalf("error should name service and field, got %v", err)
	}
}

func TestValidateServiceTemplates_EnvCannotUseName(t *testing.T) {
	svc := &config.Service{
		Name: "postgres",
		Docker: config.Docker{
			EnvVars: []string{"FOO={{.NAME}}"},
		},
	}
	err := ValidateServiceTemplates(svc)
	if err == nil {
		t.Fatal("expected error: envVars are rendered with PASSWORD only")
	}
	if !strings.Contains(err.Error(), "postgres") || !strings.Contains(err.Error(), "envVars[0]") {
		t.Fatalf("error should name service and field, got %v", err)
	}
}

func TestValidateServiceTemplates_VolumeCannotUsePassword(t *testing.T) {
	svc := &config.Service{
		Name: "postgres",
		Docker: config.Docker{
			Volumes: []config.Volume{{Source: "{{.PASSWORD}}-data", IsNamed: true}},
		},
	}
	err := ValidateServiceTemplates(svc)
	if err == nil {
		t.Fatal("expected error: volume sources are rendered with NAME only")
	}
	if !strings.Contains(err.Error(), "volumes[0].source") {
		t.Fatalf("error should name the field, got %v", err)
	}
}

func TestValidateServiceTemplates_CliCannotUseName(t *testing.T) {
	svc := &config.Service{
		Name:              "postgres",
		CliConnectCommand: "psql {{.NAME}}",
	}
	err := ValidateServiceTemplates(svc)
	if err == nil {
		t.Fatal("expected error: cliConnectCommand is rendered with PASSWORD only")
	}
	if !strings.Contains(err.Error(), "cliConnectCommand") {
		t.Fatalf("error should name the field, got %v", err)
	}
}

func TestValidateServiceTemplates_OK(t *testing.T) {
	svc := &config.Service{
		Name:              "postgres",
		CliConnectCommand: "psql {{.PASSWORD}}",
		Docker: config.Docker{
			EnvVars: []string{"POSTGRES_PASSWORD={{.PASSWORD}}"},
			Volumes: []config.Volume{{Source: "{{.NAME}}-data", IsNamed: true}},
		},
	}
	if err := ValidateServiceTemplates(svc); err != nil {
		t.Fatal(err)
	}
}

func TestReplacePassword_WithCustomPassword(t *testing.T) {
	service := &config.Service{
		Name:           "postgres",
		CustomPassword: "custom-password",
	}

	result, err := replacePassword("password={{.PASSWORD}}", service, "default-password")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result != "password=custom-password" {
		t.Errorf("Expected custom password, got '%s'", result)
	}
}

func TestReplacePassword_WithDefaultPassword(t *testing.T) {
	service := &config.Service{
		Name:           "postgres",
		CustomPassword: "",
	}

	result, err := replacePassword("password={{.PASSWORD}}", service, "default-password")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result != "password=default-password" {
		t.Errorf("Expected default password, got '%s'", result)
	}
}
