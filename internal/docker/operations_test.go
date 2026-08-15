package docker

import (
	"testing"

	"github.com/stefanjarina/sda/internal/config"
)

func TestToServiceInfo(t *testing.T) {
	config.CONFIG.Prefix = "sda"

	e := psEntry{
		ID:     "abc123",
		Names:  "sda-postgres",
		Image:  "postgres:16",
		Ports:  "0.0.0.0:5432->5432/tcp, [::]:5432->5432/tcp",
		Status: "Up 3 hours",
		State:  "running",
	}

	info := toServiceInfo(e)

	if info.Name != "postgres" {
		t.Errorf("Expected name 'postgres', got '%s'", info.Name)
	}
	if info.ContainerName != "sda-postgres" {
		t.Errorf("Expected container name 'sda-postgres', got '%s'", info.ContainerName)
	}
	if info.Version != "16" {
		t.Errorf("Expected version '16', got '%s'", info.Version)
	}
	if info.StatusIcon != "●" {
		t.Errorf("Expected running icon, got '%s'", info.StatusIcon)
	}
	if len(info.Ports) != 1 || info.Ports[0] != "5432:5432" {
		t.Errorf("Expected deduplicated port mapping, got %v", info.Ports)
	}
}

func TestToServiceInfo_StatusIcons(t *testing.T) {
	config.CONFIG.Prefix = "sda"

	tests := []struct {
		state    string
		expected string
	}{
		{"running", "●"},
		{"exited", "✗"},
		{"created", "○"},
	}

	for _, tt := range tests {
		t.Run(tt.state, func(t *testing.T) {
			info := toServiceInfo(psEntry{Names: "sda-redis", Image: "redis:latest", State: tt.state})
			if info.StatusIcon != tt.expected {
				t.Errorf("Expected icon '%s' for state '%s', got '%s'", tt.expected, tt.state, info.StatusIcon)
			}
		})
	}
}

func TestOwnedNameFilterAnchorsPrefix(t *testing.T) {
	config.CONFIG.Prefix = "sda"
	got := ownedNameFilter()
	want := "name=^sda-"
	if got != want {
		t.Fatalf("ownedNameFilter() = %q, want %q", got, want)
	}
}

func TestOwnedNameFilterQuotesMeta(t *testing.T) {
	config.CONFIG.Prefix = "sda.dev"
	got := ownedNameFilter()
	if got != "name=^sda\\.dev-" {
		t.Fatalf("ownedNameFilter() = %q, want quoted meta", got)
	}
}

func TestIsOwnedContainer(t *testing.T) {
	config.CONFIG.Prefix = "sda"
	tests := []struct {
		name string
		in   string
		want bool
	}{
		{"owned", "sda-postgres", true},
		{"leading slash", "/sda-postgres", true},
		{"substring in middle", "my-sda-staging", false},
		{"suffix only", "team-sda-runner", false},
		{"unrelated", "legacy-db", false},
		{"prefix without dash", "sda", false},
		{"empty", "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isOwnedContainer(tt.in); got != tt.want {
				t.Fatalf("isOwnedContainer(%q) = %v, want %v", tt.in, got, tt.want)
			}
		})
	}
}
