package utils

import (
	"fmt"

	"github.com/charmbracelet/huh"
)

func Confirm(question string) bool {
	confirm := true
	huh.NewConfirm().
		Title(fmt.Sprintf("%s ", question)).
		Affirmative("Yes!").
		Negative("No.").
		Value(&confirm).
		Inline(true).
		Value(&confirm).
		Run()
	return confirm
}
