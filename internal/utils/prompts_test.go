package utils

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/spf13/viper"
)

func TestConfirmJSONModeRequiresYes(t *testing.T) {
	if os.Getenv("SDA_TEST_CONFIRM_JSON") == "1" {
		viper.Reset()
		viper.Set("json", true)
		Confirm("Stop all services (postgres)? (Y/n): ")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, os.Args[0], "-test.run=TestConfirmJSONModeRequiresYes")
	cmd.Env = append(os.Environ(), "SDA_TEST_CONFIRM_JSON=1")
	out, err := cmd.CombinedOutput()
	if ctx.Err() == context.DeadlineExceeded {
		t.Fatal("Confirm() hung in JSON mode; it must refuse instead of prompting")
	}
	if err == nil {
		t.Fatal("expected process to exit non-zero")
	}
	var doc map[string]string
	if jsonErr := json.Unmarshal(out, &doc); jsonErr != nil {
		t.Fatalf("expected a single JSON error document: %v\n%s", jsonErr, out)
	}
	if !strings.Contains(doc["error"], "--json requires -y") {
		t.Fatalf("wrong error: %s", out)
	}
}
