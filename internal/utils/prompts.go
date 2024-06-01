package utils

import (
	"github.com/erikgeiser/promptkit/confirmation"
)

func Confirm(question string) bool {
	input := confirmation.New(question, confirmation.Yes)

	answer, err := input.RunPrompt()
	if err != nil {
		return false
	}

	return answer
}
