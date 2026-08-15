package utils

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/spf13/viper"
)

func captureStdout(t *testing.T, fn func()) string {
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

func TestJSONAlwaysEmitsJSON(t *testing.T) {
	viper.Reset()
	viper.Set("json", false) // the whole point: JSON() ignores the flag

	out := captureStdout(t, func() {
		JSON([]map[string]string{{"Name": "postgres"}})
	})
	var decoded []map[string]string
	if err := json.Unmarshal([]byte(out), &decoded); err != nil {
		t.Fatalf("JSON() did not emit parseable JSON: %v\n%s", err, out)
	}
	if len(decoded) != 1 || decoded[0]["Name"] != "postgres" {
		t.Fatalf("decoded %+v", decoded)
	}
}

func TestMessageWithoutJSONFlagDoesNotSwallowUnknownTypes(t *testing.T) {
	viper.Reset()
	viper.Set("json", false)

	type serviceInfo struct{ Name string }
	out := captureStdout(t, func() {
		Message([]serviceInfo{{Name: "postgres"}})
	})
	if strings.TrimSpace(out) == "" {
		t.Fatal("Message printed nothing for an unhandled type; outputText needs a default branch")
	}
	if !strings.Contains(out, "postgres") {
		t.Fatalf("expected the value to appear, got %q", out)
	}
}

func TestOutputJSONDoesNotPrintNullAfterMarshalError(t *testing.T) {
	// channels cannot be marshalled
	out := captureStdout(t, func() {
		outputJSON(make(chan int))
	})
	if strings.Count(out, "{") != 1 {
		t.Fatalf("expected a single JSON object, got %q", out)
	}
	if strings.Contains(out, "null") {
		t.Fatalf("marshal failure fell through and printed null: %q", out)
	}
}

func TestErrorJSONIsASingleDocument(t *testing.T) {
	viper.Reset()
	viper.Set("json", true)
	out := captureStdout(t, func() {
		Error("Failed to check if service 'nonexistent' exists: boom")
	})
	var doc map[string]string
	if err := json.Unmarshal([]byte(out), &doc); err != nil {
		t.Fatalf("Error() emitted more than one JSON document (or invalid JSON): %v\n%s", err, out)
	}
	if doc["error"] == "" {
		t.Fatalf("missing error field: %s", out)
	}
}

func TestErrorAndExitSkipsEmptyMessage(t *testing.T) {
	if os.Getenv("SDA_TEST_ERRORANDEXIT") == "1" {
		viper.Reset()
		viper.Set("json", true)
		ErrorAndExit("")
		return
	}
	cmd := exec.Command(os.Args[0], "-test.run=TestErrorAndExitSkipsEmptyMessage")
	cmd.Env = append(os.Environ(), "SDA_TEST_ERRORANDEXIT=1")
	out, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatal("expected process to exit non-zero")
	}
	if strings.Contains(string(out), `"error"`) || strings.TrimSpace(string(out)) != "" {
		t.Fatalf("ErrorAndExit(\"\") should emit nothing, got %q", out)
	}
}

func TestResultJSONIsASingleDocument(t *testing.T) {
	viper.Reset()
	viper.Set("json", true)
	out := captureStdout(t, func() {
		Result("Started service 'alpha'")
	})
	var doc map[string]any
	if err := json.Unmarshal([]byte(out), &doc); err != nil {
		t.Fatalf("Result() emitted more than one JSON document: %v\n%s", err, out)
	}
	if doc["ok"] != true {
		t.Fatalf("ok: %v", doc["ok"])
	}
	if doc["message"] != "Started service 'alpha'" {
		t.Fatalf("message: %v", doc["message"])
	}
}

func TestResultTextPrintsMessage(t *testing.T) {
	viper.Reset()
	viper.Set("json", false)
	out := captureStdout(t, func() {
		Result("Started service 'alpha'")
	})
	if strings.TrimSpace(out) != "Started service 'alpha'" {
		t.Fatalf("got %q", out)
	}
}

func TestProgressSilentInJSONMode(t *testing.T) {
	viper.Reset()
	viper.Set("json", true)
	out := captureStdout(t, func() {
		Progress("Pulling image '%s'...\n", "postgres:16")
	})
	if strings.TrimSpace(out) != "" {
		t.Fatalf("progress leaked onto stdout in JSON mode: %q", out)
	}
}

func TestProgressPrintsInTextMode(t *testing.T) {
	viper.Reset()
	viper.Set("json", false)
	out := captureStdout(t, func() {
		Progress("Pulling image '%s'...\n", "postgres:16")
	})
	if !strings.Contains(out, "Pulling image 'postgres:16'") {
		t.Fatalf("got %q", out)
	}
}

func TestJSONModeFollowsViper(t *testing.T) {
	viper.Reset()
	viper.Set("json", false)
	if JSONMode() {
		t.Fatal("expected false")
	}
	viper.Set("json", true)
	if !JSONMode() {
		t.Fatal("expected true")
	}
}

func TestCancelledJSONEmitsDocumentAndExitsNonZero(t *testing.T) {
	if os.Getenv("SDA_TEST_CANCELLED_JSON") == "1" {
		viper.Reset()
		viper.Set("json", true)
		Cancelled()
		return
	}
	cmd := exec.Command(os.Args[0], "-test.run=TestCancelledJSONEmitsDocumentAndExitsNonZero")
	cmd.Env = append(os.Environ(), "SDA_TEST_CANCELLED_JSON=1")
	out, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatal("expected process to exit non-zero")
	}
	var doc map[string]any
	if jsonErr := json.Unmarshal(out, &doc); jsonErr != nil {
		t.Fatalf("Cancelled() must emit one JSON document: %v\n%s", jsonErr, out)
	}
	if doc["ok"] != false {
		t.Fatalf("ok: %v", doc["ok"])
	}
	if doc["message"] != "cancelled" {
		t.Fatalf("message: %v", doc["message"])
	}
}

func TestCancelledTextEmitsNothingAndExitsZero(t *testing.T) {
	if os.Getenv("SDA_TEST_CANCELLED_TEXT") == "1" {
		viper.Reset()
		viper.Set("json", false)
		Cancelled()
		return
	}
	cmd := exec.Command(os.Args[0], "-test.run=TestCancelledTextEmitsNothingAndExitsZero")
	cmd.Env = append(os.Environ(), "SDA_TEST_CANCELLED_TEXT=1")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("text mode cancel should exit 0: %v\n%s", err, out)
	}
	if strings.TrimSpace(string(out)) != "" {
		t.Fatalf("text mode cancel should be silent, got %q", out)
	}
}

func TestRequireYesInJSONMode(t *testing.T) {
	viper.Reset()
	viper.Set("json", false)
	if err := requireYesInJSONMode(); err != nil {
		t.Fatalf("text mode should allow prompts: %v", err)
	}
	viper.Set("json", true)
	err := requireYesInJSONMode()
	if err == nil {
		t.Fatal("expected error in JSON mode")
	}
	if !strings.Contains(err.Error(), "--json requires -y") {
		t.Fatalf("wrong message: %v", err)
	}
}
