package utils

import (
	"fmt"

	"github.com/erikgeiser/promptkit/confirmation"
)

func requireYesInJSONMode() error {
	if JSONMode() {
		return fmt.Errorf("--json requires -y; confirmation prompts are not supported in JSON mode")
	}
	return nil
}

func Confirm(question string) bool {
	if err := requireYesInJSONMode(); err != nil {
		ErrorAndExit(err.Error())
	}

	input := confirmation.New(question, confirmation.Yes)

	answer, err := input.RunPrompt()
	if err != nil {
		return false
	}

	return answer
}
