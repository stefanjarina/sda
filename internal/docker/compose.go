package docker

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/docker/cli/cli/command"
	"github.com/docker/cli/cli/flags"
	"github.com/docker/compose/v5/pkg/api"
	"github.com/docker/compose/v5/pkg/compose"
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

// createComposeService creates a Docker Compose API service
func createComposeService() (api.Compose, error) {
	dockerCli, err := command.NewDockerCli()
	if err != nil {
		return nil, fmt.Errorf("failed to create Docker CLI: %w", err)
	}

	// Initialize Docker CLI with options
	dockerContext := "default"
	opts := &flags.ClientOptions{Context: dockerContext, LogLevel: "error"}
	if err := dockerCli.Initialize(opts); err != nil {
		return nil, fmt.Errorf("failed to initialize Docker CLI: %w", err)
	}

	// Create compose service
	composeService, err := compose.NewComposeService(dockerCli)
	if err != nil {
		return nil, fmt.Errorf("failed to create compose service: %w", err)
	}

	return composeService, nil
}

// ComposeUp starts a compose project
func (d *Api) ComposeUp(service config.Service, build bool, recreate bool) error {
	if service.Compose == "" {
		return fmt.Errorf("no compose file specified for service %s", service.Name)
	}

	// Resolve compose file path
	composePath, err := resolveComposePath(service)
	if err != nil {
		return err
	}

	// Create compose service
	composeService, err := createComposeService()
	if err != nil {
		return err
	}

	// Load project
	ctx := context.Background()
	project, err := composeService.LoadProject(ctx, api.ProjectLoadOptions{
		ConfigPaths: []string{composePath},
		ProjectName: service.Name,
	})
	if err != nil {
		return fmt.Errorf("failed to load compose project: %w", err)
	}

	// Build up options
	upOptions := api.UpOptions{
		Create: api.CreateOptions{
			Services:             []string{},
			RemoveOrphans:        true,
			Recreate:             api.RecreateNever,
			RecreateDependencies: api.RecreateNever,
			Build:                nil,
		},
		Start: api.StartOptions{
			Project:  project,
			AttachTo: []string{},
		},
	}

	// Apply build flag
	if build {
		upOptions.Create.Build = &api.BuildOptions{
			Services: []string{},
		}
	}

	// Apply recreate flag
	if recreate {
		upOptions.Create.Recreate = api.RecreateForce
		upOptions.Create.RecreateDependencies = api.RecreateForce
	}

	// Execute up
	return composeService.Up(ctx, project, upOptions)
}

// ComposeStart starts a compose project (similar to docker compose start)
func (d *Api) ComposeStart(service config.Service) error {
	if service.Compose == "" {
		return fmt.Errorf("no compose file specified for service %s", service.Name)
	}

	// Resolve compose file path (for validation)
	_, err := resolveComposePath(service)
	if err != nil {
		return err
	}

	// Create compose service
	composeService, err := createComposeService()
	if err != nil {
		return err
	}

	// Execute start
	ctx := context.Background()
	startOptions := api.StartOptions{
		Services: []string{},
	}

	return composeService.Start(ctx, service.Name, startOptions)
}

// ComposeStop stops a compose project without removing it
func (d *Api) ComposeStop(service config.Service) error {
	if service.Compose == "" {
		return fmt.Errorf("no compose file specified for service %s", service.Name)
	}

	// Resolve compose file path (for validation)
	_, err := resolveComposePath(service)
	if err != nil {
		return err
	}

	// Create compose service
	composeService, err := createComposeService()
	if err != nil {
		return err
	}

	// Execute stop
	ctx := context.Background()
	stopOptions := api.StopOptions{
		Services: []string{},
	}

	return composeService.Stop(ctx, service.Name, stopOptions)
}

// ComposeDown stops and removes a compose project
func (d *Api) ComposeDown(service config.Service, removeVolumes bool) error {
	if service.Compose == "" {
		return fmt.Errorf("no compose file specified for service %s", service.Name)
	}

	// Resolve compose file path (for validation)
	_, err := resolveComposePath(service)
	if err != nil {
		return err
	}

	// Create compose service
	composeService, err := createComposeService()
	if err != nil {
		return err
	}

	// Execute down
	ctx := context.Background()
	downOptions := api.DownOptions{
		RemoveOrphans: true,
		Project:       nil,
		Volumes:       removeVolumes,
		Images:        "",
	}

	return composeService.Down(ctx, service.Name, downOptions)
}

// ComposeLogConsumer implements api.LogConsumer interface for streaming logs
type ComposeLogConsumer struct{}

func (l *ComposeLogConsumer) Log(containerName, message string) {
	fmt.Printf("[%s] %s\n", containerName, message)
}

func (l *ComposeLogConsumer) Err(containerName, message string) {
	fmt.Fprintf(os.Stderr, "[%s] %s\n", containerName, message)
}

func (l *ComposeLogConsumer) Status(container, msg string) {}

// ComposeLogs shows logs from a compose project
func (d *Api) ComposeLogs(service config.Service, follow bool) error {
	if service.Compose == "" {
		return fmt.Errorf("no compose file specified for service %s", service.Name)
	}

	// Resolve compose file path (for validation)
	_, err := resolveComposePath(service)
	if err != nil {
		return err
	}

	// Create compose service
	composeService, err := createComposeService()
	if err != nil {
		return err
	}

	// Execute logs
	ctx := context.Background()
	logOptions := api.LogOptions{
		Services: []string{},
		Follow:   follow,
	}

	consumer := &ComposeLogConsumer{}
	return composeService.Logs(ctx, service.Name, consumer, logOptions)
}
