package utils

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
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
