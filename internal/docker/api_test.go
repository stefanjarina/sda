package docker

import (
	"testing"

	"github.com/stefanjarina/sda/internal/config"
)

func TestNew_NilConfig(t *testing.T) {
	_, err := New(nil)
	if err == nil {
		t.Fatal("expected error for nil config")
	}
}

func TestTwoApisIndependentPrefixes(t *testing.T) {
	a := &Api{cfg: &config.Config{Prefix: "sda"}}
	b := &Api{cfg: &config.Config{Prefix: "dev"}}
	if got := a.containerName("pg"); got != "sda-pg" {
		t.Fatalf("a.containerName = %q, want sda-pg", got)
	}
	if got := b.containerName("pg"); got != "dev-pg" {
		t.Fatalf("b.containerName = %q, want dev-pg", got)
	}
}
