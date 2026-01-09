package docker

import (
	"fmt"
	"io"
	"strings"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/stefanjarina/sda/internal/config"
	"github.com/stefanjarina/sda/internal/utils"
)

func (d *Api) ListAvailable() []ServiceInfo {
	var services []ServiceInfo

	for _, s := range config.CONFIG.Services {
		serviceInfo := &ServiceInfo{
			Name:          s.Name,
			ContainerName: fmt.Sprintf("%s-%s", config.CONFIG.Prefix, s.Name),
			ID:            "",
			Image:         s.Docker.ImageName,
			Version:       s.Version,
			Ports:         []string{},
			Status:        "",
		}
		services = append(services, *serviceInfo)
	}

	return services
}

func (d *Api) ListCreated() []ServiceInfo {
	listOptions := container.ListOptions{
		Filters: filters.NewArgs(
			filters.Arg("name", fmt.Sprintf("%s-", config.CONFIG.Prefix)),
		),
		All: true,
	}

	result, _ := d.client.ContainerList(d.ctx, listOptions)
	var services []ServiceInfo

	for _, s := range result {
		statusIcon := "○"
		if strings.HasPrefix(s.Status, "Up") {
			statusIcon = "●"
		} else if strings.HasPrefix(s.Status, "Exited") {
			statusIcon = "✗"
		}
		serviceInfo := &ServiceInfo{
			Name:          getNameFromContainerName(s.Names[0]),
			ContainerName: s.Names[0][1:],
			ID:            s.ID,
			Image:         s.Image,
			Version:       getVersionFromImageName(s.Image),
			Ports:         getPortsFromContainer(s.Ports),
			Status:        s.Status,
			StatusIcon:    statusIcon,
		}
		services = append(services, *serviceInfo)
	}

	return services
}

func (d *Api) ListRunning() []ServiceInfo {
	listOptions := container.ListOptions{
		Filters: filters.NewArgs(
			filters.Arg("name", fmt.Sprintf("%s-", config.CONFIG.Prefix)),
			filters.Arg("status", "running"),
		),
		All: true,
	}

	result, _ := d.client.ContainerList(d.ctx, listOptions)
	var services []ServiceInfo

	for _, s := range result {
		serviceInfo := &ServiceInfo{
			Name:          getNameFromContainerName(s.Names[0]),
			ContainerName: s.Names[0][1:],
			ID:            s.ID,
			Image:         s.Image,
			Version:       getVersionFromImageName(s.Image),
			Ports:         getPortsFromContainer(s.Ports),
			Status:        s.Status,
			StatusIcon:    "●",
		}
		services = append(services, *serviceInfo)
	}

	return services
}

func (d *Api) ListStopped() []ServiceInfo {
	listOptions := container.ListOptions{
		Filters: filters.NewArgs(
			filters.Arg("name", fmt.Sprintf("%s-", config.CONFIG.Prefix)),
			filters.Arg("status", "exited"),
		),
		All: true,
	}

	result, _ := d.client.ContainerList(d.ctx, listOptions)
	var services []ServiceInfo

	for _, s := range result {
		serviceInfo := &ServiceInfo{
			Name:          getNameFromContainerName(s.Names[0]),
			ContainerName: s.Names[0][1:],
			ID:            s.ID,
			Image:         s.Image,
			Version:       getVersionFromImageName(s.Image),
			Ports:         getPortsFromContainer(s.Ports),
			Status:        s.Status,
			StatusIcon:    "✗",
		}
		services = append(services, *serviceInfo)
	}

	return services
}

func (d *Api) GetInfo(name string) ServiceInfo {
	listOptions := container.ListOptions{
		Filters: filters.NewArgs(
			filters.Arg("name", fmt.Sprintf("%s-%s", config.CONFIG.Prefix, name)),
		),
		All: true,
	}

	result, _ := d.client.ContainerList(d.ctx, listOptions)
	s := result[0]

	serviceInfo := ServiceInfo{
		Name:          getNameFromContainerName(s.Names[0]),
		ContainerName: s.Names[0][1:],
		ID:            s.ID,
		Image:         s.Image,
		Version:       getVersionFromImageName(s.Image),
		Ports:         getPortsFromContainer(s.Ports),
		Status:        s.Status,
	}

	return serviceInfo
}

func (d *Api) Start(name string) error {
	err := d.client.ContainerStart(d.ctx, fmt.Sprintf("%s-%s", config.CONFIG.Prefix, name), container.StartOptions{})
	return err
}

func (d *Api) Stop(name string) error {
	err := d.client.ContainerStop(d.ctx, fmt.Sprintf("%s-%s", config.CONFIG.Prefix, name), container.StopOptions{})
	return err
}

func (d *Api) Remove(name string) error {
	err := d.client.ContainerRemove(d.ctx, fmt.Sprintf("%s-%s", config.CONFIG.Prefix, name), container.RemoveOptions{
		Force:         true,
		RemoveVolumes: true,
	})

	return err
}

func (d *Api) RemoveVolumes(names []string) {
	for _, name := range names {
		_ = d.client.VolumeRemove(d.ctx, fmt.Sprintf("%s-%s", config.CONFIG.Prefix, name), true)
	}
}

func (d *Api) Connect(name string, customPassword string, web bool) error {
	service := config.CONFIG.GetServiceByName(name)

	if web {
		return handleWebConnect(service)
	} else {
		return handleCliConnect(service, customPassword, name)
	}
}

func (d *Api) Exists(name string) bool {
	listOptions := container.ListOptions{
		Filters: filters.NewArgs(
			filters.Arg("name", fmt.Sprintf("%s-%s", config.CONFIG.Prefix, name)),
		),
		All: true,
	}

	result, _ := d.client.ContainerList(d.ctx, listOptions)
	return len(result) > 0
}

func handleWebConnect(service *config.Service) error {
	url := service.WebConnectUrl
	err := utils.OpenURL(url)
	if err != nil {
		return err
	}

	return nil
}

func handleCliConnect(service *config.Service, customPassword, name string) error {
	var cmd string
	if service.HasPassword {
		var passwordToUse string
		if customPassword != "" {
			passwordToUse = customPassword
		} else {
			passwordToUse = config.CONFIG.Password
		}
		cmd = replacePassword(service.CliConnectCommand, service, passwordToUse)
	} else {
		cmd = service.CliConnectCommand
	}

	err := utils.RunInteractive(cmd, fmt.Sprintf("%s-%s", config.CONFIG.Prefix, name))
	if err != nil {
		return err
	}

	return nil
}

func (d *Api) GetContainerLogs(name string, options container.LogsOptions) (io.ReadCloser, error) {
	containerName := fmt.Sprintf("%s-%s", config.CONFIG.Prefix, name)
	return d.client.ContainerLogs(d.ctx, containerName, options)
}
