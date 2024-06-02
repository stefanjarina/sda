package docker

import (
	"fmt"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/volume"
	"github.com/docker/go-connections/nat"
	"github.com/stefanjarina/sda/internal/config"
	"io"
)

func (d *Api) Create(name string) error {
	service := config.CONFIG.GetServiceByName(name)

	if d.Exists(name) {
		return fmt.Errorf("Service %s already exists", name)
	}

	err := d.fetchImageIfNotExists(service.Docker.ImageName, service.Version)
	if err != nil {
		return err
	}

	containerConfig := &container.Config{}
	hostConfig := &container.HostConfig{}

	containerConfig.Image = fmt.Sprintf("%s:%s", service.Docker.ImageName, service.Version)

	if service.Docker.EnvVars != nil {
		// patchEnvVars
		var envVars []string
		for _, envVar := range service.Docker.EnvVars {
			envVars = append(envVars, replacePassword(envVar, service, config.CONFIG.Password))
		}

		containerConfig.Env = envVars
	}

	if service.Docker.CustomAppCommands != nil {
		// patchEnvVars
		var customCommands []string
		for _, customCommand := range service.Docker.CustomAppCommands {
			customCommands = append(customCommands, replacePassword(customCommand, service, config.CONFIG.Password))
		}

		containerConfig.Cmd = customCommands
	}

	if service.Docker.Volumes != nil {
		//hostConfig.Mounts, _ = d.mapMounts(service.Docker.Volumes, fmt.Sprintf("%s-%s", config.CONFIG.Prefix, name))
		containerConfig.Volumes = mapVolumes(service.Docker.Volumes)
		hostConfig.Binds = mapBinds(service.Docker.Volumes, fmt.Sprintf("%s-%s", config.CONFIG.Prefix, name))
	}

	if service.Docker.PortMappings != nil {
		ports, _ := mapPorts(service.Docker.PortMappings)
		hostConfig.PortBindings = ports
	}

	//parse --ulimit nofile=262144:262144 from additional arguments
	hostConfig.Ulimits = parseUlimits(service.Docker.AdditionalDockerArguments)

	_, err = d.client.ContainerCreate(d.ctx, containerConfig, hostConfig, nil, nil, fmt.Sprintf("%s-%s", config.CONFIG.Prefix, name))
	if err != nil {
		return err
	}

	return nil
}

func (d *Api) fetchImageIfNotExists(name string, version string) error {
	nameAndVersion := fmt.Sprintf("%s:%s", name, version)
	_, _, err := d.client.ImageInspectWithRaw(d.ctx, name)
	if err != nil {
		reader, err := d.client.ImagePull(d.ctx, nameAndVersion, image.PullOptions{})
		if err != nil {
			return err
		}
		defer reader.Close()

		fmt.Printf("Unable to find image '%s' locally. Pulling '%s' from '%s'...\n", nameAndVersion, version, name)
		_, err = io.Copy(io.Discard, reader)
		if err != nil {
			return err
		}
	}
	return nil
}

func (d *Api) createDockerVolume(volumeName string) error {
	volumeOptions := volume.CreateOptions{
		Name: volumeName,
	}

	_, err := d.client.VolumeCreate(d.ctx, volumeOptions)
	if err != nil {
		return err
	}

	return nil
}

func (d *Api) mapMounts(volumes []config.Volume, name string) ([]mount.Mount, error) {
	var mounts []mount.Mount
	for _, v := range volumes {
		volumeSource := replacePlaceholder(v.Source, map[string]string{"NAME": name})
		if v.IsNamed {
			err := d.createDockerVolume(volumeSource)
			if err != nil {
				return nil, err
			}
		}

		var mountType mount.Type = mount.TypeBind
		if v.IsNamed {
			mountType = mount.TypeVolume
		}

		m := mount.Mount{
			Type:   mountType,
			Source: volumeSource,
			Target: v.Target,
		}
		mounts = append(mounts, m)
	}
	return mounts, nil
}

func mapVolumes(volumes []config.Volume) map[string]struct{} {
	result := make(map[string]struct{})
	for _, v := range volumes {
		result[v.Target] = struct{}{}
	}
	return result
}

func mapBinds(volumes []config.Volume, name string) []string {
	var result []string
	for _, v := range volumes {
		result = append(result, fmt.Sprintf("%s:%s", replacePlaceholder(v.Source, map[string]string{"NAME": name}), v.Target))
	}
	return result
}

func mapPorts(ports []config.PortMapping) (nat.PortMap, error) {
	portBindings := nat.PortMap{}
	for _, port := range ports {
		portBinding := nat.PortBinding{
			HostIP:   "",
			HostPort: fmt.Sprintf("%d", port.Host),
		}

		var portName nat.Port = nat.Port(fmt.Sprintf("%d/tcp", port.Container))
		portBindings[portName] = []nat.PortBinding{portBinding}
	}
	return portBindings, nil
}
