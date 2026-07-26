package docker

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/stefanjarina/sda/internal/config"
)

// resolveComposePath resolves the compose file path
// - If path is absolute, use as-is
// - If path is relative, resolve relative to config directory (~/.config/sda)
// - If path is a directory, search for docker-compose.yaml or docker-compose.yml
// - If path is a file, use directly
func resolveComposePath(service config.Service) (string, error) {
	composePath := service.Compose

	// Expand path relative to config directory if not absolute
	if !filepath.IsAbs(composePath) {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("failed to get user home directory: %w", err)
		}
		configDir := filepath.Join(home, ".config", "sda")
		composePath = filepath.Join(configDir, composePath)
	}

	// Check if path exists
	info, err := os.Stat(composePath)
	if os.IsNotExist(err) {
		return "", fmt.Errorf("compose path not found: %s", composePath)
	}
	if err != nil {
		return "", fmt.Errorf("failed to stat compose path: %w", err)
	}

	// If it's a directory, search for compose file
	if info.IsDir() {
		// Try docker-compose.yaml first
		yamlPath := filepath.Join(composePath, "docker-compose.yaml")
		if _, err := os.Stat(yamlPath); err == nil {
			composePath = yamlPath
		} else {
			// Try docker-compose.yml
			ymlPath := filepath.Join(composePath, "docker-compose.yml")
			if _, err := os.Stat(ymlPath); err == nil {
				composePath = ymlPath
			} else {
				return "", fmt.Errorf("no docker-compose.yaml or docker-compose.yml found in directory: %s", composePath)
			}
		}
	}

	// Validate folder name matches service name
	folderName := filepath.Base(filepath.Dir(composePath))
	if folderName != service.Name {
		return "", fmt.Errorf("folder name '%s' must match service name '%s'", folderName, service.Name)
	}

	return composePath, nil
}

// composeArgs returns the leading `docker compose` argv shared by every
// compose subcommand: the resolved compose file and the project name.
func (d *Api) composeArgs(service config.Service) ([]string, error) {
	if service.Compose == "" {
		return nil, fmt.Errorf("no compose file specified for service %s", service.Name)
	}

	composePath, err := resolveComposePath(service)
	if err != nil {
		return nil, err
	}

	return []string{"compose", "-f", composePath, "-p", service.Name}, nil
}

// ComposeUp starts a compose project
func (d *Api) ComposeUp(service config.Service, build bool, recreate bool) error {
	args, err := d.composeArgs(service)
	if err != nil {
		return err
	}

	args = append(args, "up", "--detach", "--remove-orphans")
	if build {
		args = append(args, "--build")
	}
	if recreate {
		args = append(args, "--force-recreate")
	}

	return d.run(args...)
}

// ComposeStart starts a compose project (similar to docker compose start)
func (d *Api) ComposeStart(service config.Service) error {
	args, err := d.composeArgs(service)
	if err != nil {
		return err
	}

	return d.run(append(args, "start")...)
}

// ComposeStop stops a compose project without removing it
func (d *Api) ComposeStop(service config.Service) error {
	args, err := d.composeArgs(service)
	if err != nil {
		return err
	}

	return d.run(append(args, "stop")...)
}

// ComposeDown stops and removes a compose project
func (d *Api) ComposeDown(service config.Service, removeVolumes bool) error {
	args, err := d.composeArgs(service)
	if err != nil {
		return err
	}

	args = append(args, "down", "--remove-orphans")
	if removeVolumes {
		args = append(args, "--volumes")
	}

	return d.run(args...)
}

// ComposeLogs shows logs from a compose project
func (d *Api) ComposeLogs(service config.Service, follow bool) error {
	args, err := d.composeArgs(service)
	if err != nil {
		return err
	}

	args = append(args, "logs")
	if follow {
		args = append(args, "--follow")
	}

	return d.run(args...)
}
