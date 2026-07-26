package docker

import (
	"fmt"
	"strings"

	"github.com/stefanjarina/sda/internal/config"
)

func (d *Api) Create(name string) error {
	service := config.CONFIG.GetServiceByName(name)

	exists, err := d.Exists(name)
	if err != nil {
		return fmt.Errorf("failed to check if service exists: %w", err)
	}
	if exists {
		return fmt.Errorf("service %s already exists", name)
	}

	imageRef := fmt.Sprintf("%s:%s", service.Docker.ImageName, service.Version)
	if err := d.fetchImageIfNotExists(imageRef); err != nil {
		return err
	}

	args := buildCreateArgs(service, containerName(name), config.CONFIG.Network, config.CONFIG.Password)

	if _, err := d.capture(args...); err != nil {
		return err
	}

	return nil
}

// buildCreateArgs builds the argv for `docker create`, translating the
// service's config into the equivalent CLI flags.
func buildCreateArgs(service *config.Service, containerName, network, password string) []string {
	args := []string{"create", "--name", containerName}

	if network != "" {
		args = append(args, "--network", network)
	}

	for _, envVar := range service.Docker.EnvVars {
		args = append(args, "--env", replacePassword(envVar, service, password))
	}

	for _, v := range service.Docker.Volumes {
		source := replacePlaceholder(v.Source, map[string]string{"NAME": containerName})
		args = append(args, "--volume", fmt.Sprintf("%s:%s", source, v.Target))
	}

	for _, p := range service.Docker.PortMappings {
		args = append(args, "--publish", fmt.Sprintf("%d:%d", p.Host, p.Container))
	}

	// e.g. "--ulimit nofile=262144:262144" - passed straight through so any
	// docker create flag works here, not just --ulimit.
	for _, extra := range service.Docker.AdditionalDockerArguments {
		args = append(args, strings.Fields(extra)...)
	}

	args = append(args, fmt.Sprintf("%s:%s", service.Docker.ImageName, service.Version))

	for _, cmd := range service.Docker.CustomAppCommands {
		args = append(args, strings.Fields(cmd)...)
	}

	return args
}

// fetchImageIfNotExists pulls imageRef (e.g. "postgres:16") unless it is
// already present locally.
func (d *Api) fetchImageIfNotExists(imageRef string) error {
	if _, err := d.capture("image", "inspect", imageRef); err == nil {
		return nil
	}

	fmt.Printf("Pulling image '%s'...\n", imageRef)
	return d.run("pull", imageRef)
}
