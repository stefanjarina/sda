package cmd

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/viper"
)

func TestClassifyConfigReadError(t *testing.T) {
	tests := []struct {
		name         string
		err          error
		userSupplied bool
		wantAction   configReadAction
		wantMsgSub   string
	}{
		{
			name:         "default path missing via ConfigFileNotFoundError",
			err:          viper.ConfigFileNotFoundError{},
			userSupplied: false,
			wantAction:   configWriteDefaults,
		},
		{
			name:         "default path missing via ErrNotExist",
			err:          os.ErrNotExist,
			userSupplied: false,
			wantAction:   configWriteDefaults,
		},
		{
			name:         "user --config path missing",
			err:          os.ErrNotExist,
			userSupplied: true,
			wantAction:   configFail,
			wantMsgSub:   "not found",
		},
		{
			name:         "yaml syntax error on default path",
			err:          errors.New("While parsing config: yaml: line 3: did not find expected key"),
			userSupplied: false,
			wantAction:   configFail,
			wantMsgSub:   "Failed to read config",
		},
		{
			name:         "permission denied on default path",
			err:          os.ErrPermission,
			userSupplied: false,
			wantAction:   configFail,
			wantMsgSub:   "Failed to read config",
		},
		{
			name:         "yaml syntax error on --config path",
			err:          errors.New("While parsing config: yaml: unmarshal errors"),
			userSupplied: true,
			wantAction:   configFail,
			wantMsgSub:   "Failed to read config",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			action, msg := classifyConfigReadError(tt.err, tt.userSupplied)
			if action != tt.wantAction {
				t.Fatalf("action = %v, want %v (msg=%q)", action, tt.wantAction, msg)
			}
			if tt.wantMsgSub != "" && !strings.Contains(msg, tt.wantMsgSub) {
				t.Fatalf("msg %q does not contain %q", msg, tt.wantMsgSub)
			}
			if tt.wantAction == configWriteDefaults && msg != "" {
				t.Fatalf("write-defaults should not carry a user-facing error, got %q", msg)
			}
		})
	}
}

func TestSaveConfigDoesNotRunOnBrokenYAML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "broken.yaml")
	original := []byte("prefix: sda\nservices:\n  - name: foo\n   badindent: true\n")
	if err := os.WriteFile(path, original, 0644); err != nil {
		t.Fatal(err)
	}

	viper.Reset()
	viper.SetConfigFile(path)
	err := viper.ReadInConfig()
	if err == nil {
		t.Fatal("expected parse error")
	}
	action, _ := classifyConfigReadError(err, true)
	if action != configFail {
		t.Fatalf("broken file classified as %v, want configFail", action)
	}

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, original) {
		t.Fatalf("config file was modified; original %d bytes, now %d", len(original), len(got))
	}
}

func TestBareJSONEnvDoesNotEnableJSONMode(t *testing.T) {
	t.Setenv("JSON", "true")
	t.Setenv("SDA_JSON", "")
	viper.Reset()
	viper.AutomaticEnv() // current production behaviour, no prefix
	if !viper.GetBool("json") {
		t.Skip("this environment does not bind a bare JSON var; nothing to regress")
	}

	viper.Reset()
	viper.SetEnvPrefix("SDA")
	viper.AutomaticEnv()
	if viper.GetBool("json") {
		t.Fatal("bare JSON=true must not enable json mode once SetEnvPrefix(\"SDA\") is set")
	}
}

func TestSDAJSONEnvEnablesJSONMode(t *testing.T) {
	t.Setenv("SDA_JSON", "true")
	viper.Reset()
	viper.SetEnvPrefix("SDA")
	viper.AutomaticEnv()
	if !viper.GetBool("json") {
		t.Fatal("SDA_JSON=true should enable json mode")
	}
}
