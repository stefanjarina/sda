package docker

import (
	"fmt"
	"github.com/docker/docker/api/types"
	"sda/internal/config"
	"strings"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
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
		serviceInfo := &ServiceInfo{
			Name:          getNameFromContainerName(s.Names[0]),
			ContainerName: s.Names[0][1:],
			ID:            s.ID,
			Image:         s.Image,
			Version:       getVersionFromImageName(s.Image),
			Ports:         getPortsFromContainer(s.Ports),
			Status:        s.Status,
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
		}
		services = append(services, *serviceInfo)
	}

	return services
}

func (d *Api) Start(name string) error {
	err := d.client.ContainerStart(d.ctx, fmt.Sprintf("%s-%s", config.CONFIG.Prefix, name), container.StartOptions{})
	return err
}

func (d *Api) Stop(name string) error {
	err := d.client.ContainerStop(d.ctx, fmt.Sprintf("%s-%s", config.CONFIG.Prefix, name), container.StopOptions{})
	return err
}

func (d *Api) Remove(name string, removeVolumes bool) error {
	err := d.client.ContainerRemove(d.ctx, fmt.Sprintf("%s-%s", config.CONFIG.Prefix, name), container.RemoveOptions{
		Force:         true,
		RemoveVolumes: removeVolumes,
	})
	return err
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

func getNameFromContainerName(containerName string) string {
	containerName = containerName[len(config.CONFIG.Prefix)+2:]
	return containerName
}

func getVersionFromImageName(imageName string) string {
	imageName = imageName[strings.LastIndex(imageName, ":")+1:]
	return imageName
}

func getPortsFromContainer(containerPorts []types.Port) []string {
	var ports []string
	for _, port := range containerPorts {
		ports = append(ports, fmt.Sprintf("%d:%d", port.PublicPort, port.PrivatePort))
	}
	return ports
}
