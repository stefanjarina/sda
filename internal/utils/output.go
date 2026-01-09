package utils

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/viper"
)

func Message(obj any) {
	if viper.GetBool("json") {
		outputJSON(obj)
	} else {
		outputText(obj)
	}
}

func ErrorAndExit(msg string) {
	Error(msg)
	os.Exit(1)
}

func Error(msg string) {
	if viper.GetBool("json") {
		outputJSON(map[string]string{"error": msg})
	} else {
		_, _ = fmt.Fprintln(os.Stderr, msg)
	}
}

func outputJSON(obj any) {
	jsonString, err := json.MarshalIndent(obj, "", "\t")
	if err != nil {
		fmt.Println("{ \"error\": \"Error marshalling JSON\" }")
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
	}
}
