package docker

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// capture runs `docker <args...>` and returns its trimmed stdout. If the
// command fails, the error includes any output written to stderr.
func (d *Api) capture(args ...string) (string, error) {
	cmd := exec.Command(d.path, args...)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", formatExecError(args, stderr.String(), err)
	}

	return strings.TrimSpace(stdout.String()), nil
}

// run runs `docker <args...>` with stdout/stderr attached to sda's own, so
// Docker's native progress output (image pulls, compose) is shown live.
func (d *Api) run(args ...string) error {
	cmd := exec.Command(d.path, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return formatExecError(args, "", err)
	}

	return nil
}

// runInteractive is run with stdin also attached, for interactive commands
// such as `docker exec -it`.
func (d *Api) runInteractive(args ...string) error {
	cmd := exec.Command(d.path, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return formatExecError(args, "", err)
	}

	return nil
}

func formatExecError(args []string, stderr string, err error) error {
	stderr = strings.TrimSpace(stderr)
	if stderr == "" {
		return fmt.Errorf("docker %s: %w", strings.Join(args, " "), err)
	}
	return fmt.Errorf("docker %s: %s", strings.Join(args, " "), stderr)
}
