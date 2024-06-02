package docker

import (
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/stefanjarina/sda/internal/config"
)

func (d *Api) CheckNetwork() bool {
	network, err := d.client.NetworkInspect(d.ctx, config.CONFIG.Network, types.NetworkInspectOptions{})
	if network.Name == "" {
		return false
	}

	return err == nil
}

func (d *Api) CreateNetwork() error {
	response, err := d.client.NetworkCreate(d.ctx, config.CONFIG.Network, types.NetworkCreate{})
	if err != nil {
		return err
	}

	fmt.Printf("Network '%s' created with ID: %s\n", config.CONFIG.Network, response.ID)

	return nil
}
