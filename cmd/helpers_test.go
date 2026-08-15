package cmd

import (
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stefanjarina/sda/internal/config"
)

func TestSelectListMode(t *testing.T) {
	tests := []struct {
		name                        string
		available, created, running bool
		stopped, compose            bool
		want                        listMode
		wantErr                     bool
	}{
		{name: "default (no flags)", want: listRunning},
		{name: "running only", running: true, want: listRunning},
		{name: "stopped only", stopped: true, want: listStopped},
		{name: "created only", created: true, want: listCreated},
		{name: "available only", available: true, want: listAvailable},
		{name: "compose only", compose: true, want: listCompose},
		{name: "running+stopped", running: true, stopped: true, wantErr: true},
		{name: "available+created", available: true, created: true, wantErr: true},
		{name: "compose+running", running: true, compose: true, wantErr: true},
		{name: "all five", available: true, created: true, running: true, stopped: true, compose: true, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := selectListMode(tt.available, tt.created, tt.running, tt.stopped, tt.compose)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Fatalf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestWantsJSON(t *testing.T) {
	newCmd := func(withFormat bool, format string) *cobra.Command {
		c := &cobra.Command{Use: "x"}
		if withFormat {
			c.Flags().String("format", "table", "")
			if format != "table" {
				_ = c.Flags().Set("format", format)
			}
		}
		return c
	}

	tests := []struct {
		name       string
		jsonFlag   bool
		withFormat bool
		format     string
		want       bool
	}{
		{name: "neither", want: false},
		{name: "global json", jsonFlag: true, want: true},
		{name: "format json", withFormat: true, format: "json", want: true},
		{name: "format table", withFormat: true, format: "table", want: false},
		{name: "both", jsonFlag: true, withFormat: true, format: "json", want: true},
		{name: "no format flag, json off", withFormat: false, want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			viper.Reset()
			viper.Set("json", tt.jsonFlag)
			got := wantsJSON(newCmd(tt.withFormat, tt.format))
			if got != tt.want {
				t.Fatalf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRequireRecreateForVolumes(t *testing.T) {
	if err := requireRecreateForVolumes(false, true); err == nil {
		t.Fatal("expected error")
	}
	if err := requireRecreateForVolumes(true, true); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := requireRecreateForVolumes(false, false); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := requireRecreateForVolumes(true, false); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRejectNoOpBulk(t *testing.T) {
	tests := []struct {
		verb             string
		running, stopped bool
		wantErr          bool
	}{
		{verb: "start", running: true, wantErr: true},
		{verb: "start", stopped: true},
		{verb: "start"},
		{verb: "stop", stopped: true, wantErr: true},
		{verb: "stop", running: true},
		{verb: "stop"},
		{verb: "remove", running: true},
		{verb: "remove", stopped: true},
	}
	for _, tt := range tests {
		err := rejectNoOpBulk(tt.verb, tt.running, tt.stopped)
		if tt.wantErr && err == nil {
			t.Fatalf("%s running=%v stopped=%v: expected error", tt.verb, tt.running, tt.stopped)
		}
		if !tt.wantErr && err != nil {
			t.Fatalf("%s running=%v stopped=%v: unexpected %v", tt.verb, tt.running, tt.stopped, err)
		}
	}
}

func TestLookupConfiguredService(t *testing.T) {
	config.CONFIG = config.Config{
		Services: []config.Service{{Name: "postgres"}},
	}
	t.Cleanup(func() { config.CONFIG = config.Config{} })

	svc, err := lookupConfiguredService("postgres")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if svc == nil || svc.Name != "postgres" {
		t.Fatalf("got %#v", svc)
	}

	_, err = lookupConfiguredService("ghost")
	if err == nil {
		t.Fatal("expected error for unknown service")
	}
	if !strings.Contains(err.Error(), "ghost") || !strings.Contains(err.Error(), "not in the list of available services") {
		t.Fatalf("wrong message: %v", err)
	}
}
