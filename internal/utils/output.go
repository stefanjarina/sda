package utils

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/viper"
)

func JSONMode() bool {
	return viper.GetBool("json")
}

func Result(msg string) {
	if JSONMode() {
		outputJSON(map[string]any{"ok": true, "message": msg})
		return
	}
	fmt.Println(msg)
}

func Progress(format string, args ...any) {
	if JSONMode() {
		return
	}
	fmt.Printf(format, args...)
}

func Message(obj any) {
	if JSONMode() {
		outputJSON(obj)
	} else {
		outputText(obj)
	}
}

func JSON(obj any) {
	outputJSON(obj)
}

func ErrorAndExit(msg string) {
	if msg != "" {
		Error(msg)
	}
	os.Exit(1)
}

// Cancelled exits after a declined confirmation. JSON mode emits one
// document and exits 1 so a consumer does not treat silence as success.
// Text mode stays silent and exits 0, matching the previous behaviour.
func Cancelled() {
	if JSONMode() {
		outputJSON(map[string]any{"ok": false, "message": "cancelled"})
		os.Exit(1)
	}
	os.Exit(0)
}

func Error(msg string) {
	if JSONMode() {
		outputJSON(map[string]string{"error": msg})
	} else {
		_, _ = fmt.Fprintln(os.Stderr, msg)
	}
}

func outputJSON(obj any) {
	jsonString, err := json.MarshalIndent(obj, "", "\t")
	if err != nil {
		fmt.Println("{ \"error\": \"Error marshalling JSON\" }")
		return
	}
	fmt.Println(string(jsonString))
}

func outputText(obj any) {
	switch obj := obj.(type) {
	case []string:
		for _, line := range obj {
			fmt.Println(line)
		}
	case string:
		fmt.Println(obj)
	default:
		fmt.Printf("%+v\n", obj)
	}
}
