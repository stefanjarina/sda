package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stefanjarina/sda/internal/config"
)

func captureCmdStdout(t *testing.T, fn func()) string {
	t.Helper()
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	old := os.Stdout
	os.Stdout = w
	defer func() { os.Stdout = old }()

	fn()
	_ = w.Close()
	var buf bytes.Buffer
	if _, err := io.Copy(&buf, r); err != nil {
		t.Fatal(err)
	}
	return buf.String()
}

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

func TestBulkSelectorValidate(t *testing.T) {
	tests := []struct {
		name    string
		sel     bulkSelector
		wantErr bool
	}{
		{name: "none", sel: bulkSelector{}},
		{name: "all", sel: bulkSelector{all: true}},
		{name: "running", sel: bulkSelector{running: true}},
		{name: "stopped", sel: bulkSelector{stopped: true}},
		{name: "all+running", sel: bulkSelector{all: true, running: true}, wantErr: true},
		{name: "running+stopped", sel: bulkSelector{running: true, stopped: true}, wantErr: true},
		{name: "all three", sel: bulkSelector{all: true, running: true, stopped: true}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.sel.validate()
			if tt.wantErr && err == nil {
				t.Fatal("expected error")
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("unexpected %v", err)
			}
			if tt.wantErr && !strings.Contains(err.Error(), "Only one of --all, --running, or --stopped") {
				t.Fatalf("wrong message: %v", err)
			}
		})
	}
}

func TestBulkPrompt(t *testing.T) {
	got := bulkPrompt("Start", "all services", []string{"alpha", "beta"}, false)
	want := "Start all services (alpha, beta)? (Y/n): "
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
	got = bulkPrompt("Remove", "all services", []string{"alpha"}, true)
	want = "Remove all services (alpha) and all volumes? (Y/n): "
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestBulkOrExactArgs(t *testing.T) {
	cmd := &cobra.Command{Use: "start"}
	cmd.Flags().Bool("all", false, "")
	cmd.Flags().Bool("running", false, "")
	cmd.Flags().Bool("stopped", false, "")

	if err := bulkOrExactArgs(cmd, []string{"postgres"}); err != nil {
		t.Fatalf("single service: %v", err)
	}
	if err := bulkOrExactArgs(cmd, []string{}); err == nil {
		t.Fatal("no args and no bulk flag should fail ExactArgs(1)")
	}
	_ = cmd.Flags().Set("all", "true")
	if err := bulkOrExactArgs(cmd, []string{}); err != nil {
		t.Fatalf("--all with no args: %v", err)
	}
	if err := bulkOrExactArgs(cmd, []string{"postgres"}); err == nil {
		t.Fatal("--all with a name should fail")
	}
}

func TestBulkDocumentAllOK(t *testing.T) {
	o := newBulkOutcome("start", "Started")
	o.record("alpha", nil)
	o.record("beta", nil)
	doc := o.document()
	if len(doc.OK) != 2 || doc.OK[0] != "alpha" || len(doc.Failed) != 0 {
		t.Fatalf("%+v", doc)
	}
	if doc.Failed == nil {
		t.Fatal("failed must be empty slice, not nil")
	}
}

func TestBulkDocumentPartialFailure(t *testing.T) {
	o := newBulkOutcome("start", "Started")
	o.record("alpha", nil)
	o.record("beta", fmt.Errorf("boom"))
	doc := o.document()
	if len(doc.OK) != 1 || doc.OK[0] != "alpha" {
		t.Fatalf("ok: %+v", doc.OK)
	}
	if len(doc.Failed) != 1 || doc.Failed[0].Service != "beta" || doc.Failed[0].Error != "boom" {
		t.Fatalf("failed: %+v", doc.Failed)
	}
}

func TestBulkDocumentEmptySlicesAreEmptyNotNil(t *testing.T) {
	doc := newBulkOutcome("start", "Started").document()
	if doc.OK == nil || doc.Failed == nil {
		t.Fatalf("nil slices: %+v", doc)
	}
	raw, err := json.Marshal(doc)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(raw), "null") {
		t.Fatalf("null in %s", raw)
	}
}

func TestBulkDocumentWarningsOmittedWhenEmpty(t *testing.T) {
	o := newBulkOutcome("remove", "Removed")
	o.record("postgres", nil)
	raw, _ := json.Marshal(o.document())
	if strings.Contains(string(raw), "warnings") {
		t.Fatalf("empty warnings should be omitted: %s", raw)
	}
	o.warn("Failed to remove volumes: boom")
	raw, _ = json.Marshal(o.document())
	if !strings.Contains(string(raw), "Failed to remove volumes: boom") {
		t.Fatalf("missing warning: %s", raw)
	}
}

func TestBulkRecordSilentInJSONMode(t *testing.T) {
	viper.Reset()
	viper.Set("json", true)
	o := newBulkOutcome("start", "Started")
	out := captureCmdStdout(t, func() {
		o.record("alpha", nil)
		o.record("beta", fmt.Errorf("boom"))
	})
	if strings.TrimSpace(out) != "" {
		t.Fatalf("record() must not print in JSON mode: %q", out)
	}
}
