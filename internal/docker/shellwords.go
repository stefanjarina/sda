package docker

import "strings"

// splitArgs splits a command string into argv tokens, respecting single and
// double quotes (which are stripped from the output). Unlike a POSIX shell,
// quotes need not be preceded by whitespace, so an adjacent quoted value like
// -p'{{.PASSWORD}}' is joined into a single token: -p{{.PASSWORD}}.
func splitArgs(s string) []string {
	var args []string
	var current strings.Builder
	hasToken := false
	var quote rune

	for _, r := range s {
		switch {
		case quote != 0:
			if r == quote {
				quote = 0
			} else {
				current.WriteRune(r)
			}
		case r == '\'' || r == '"':
			quote = r
			hasToken = true
		case r == ' ' || r == '\t':
			if hasToken {
				args = append(args, current.String())
				current.Reset()
				hasToken = false
			}
		default:
			current.WriteRune(r)
			hasToken = true
		}
	}
	if hasToken {
		args = append(args, current.String())
	}

	return args
}
