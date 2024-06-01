package docker

import (
	"fmt"
	"sda/internal/config"
	"sda/internal/utils"

	"github.com/docker/docker/api/types"
)

func (d *Api) CheckNetwork() bool {
	network, err := d.client.NetworkInspect(d.ctx, config.CONFIG.Network, types.NetworkInspectOptions{})
	if network.Name == "" {
		return false
	}

	return err == nil
}

func (d *Api) CreateNetwork() error {
	fmt.Printf("Network '%s' does not exist, create? (Y/n)", config.CONFIG.Network)

	answer := utils.Confirm(fmt.Sprintf("Network '%s' does not exist, create?", config.CONFIG.Network))

	if answer {
		response, err := d.client.NetworkCreate(d.ctx, config.CONFIG.Network, types.NetworkCreate{})
		if err != nil {
			return err
		}

		fmt.Printf("Network '%s' created with ID: %s\n", config.CONFIG.Network, response.ID)
	}

	return nil
}
