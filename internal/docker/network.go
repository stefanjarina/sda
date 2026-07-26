package docker

import (
	"fmt"

	"github.com/stefanjarina/sda/internal/config"
)

func (d *Api) CheckNetwork() bool {
	_, err := d.capture("network", "inspect", config.CONFIG.Network)
	return err == nil
}

func (d *Api) CreateNetwork() error {
	id, err := d.capture("network", "create", config.CONFIG.Network)
	if err != nil {
		return err
	}

	fmt.Printf("Created network '%s' with ID: %s\n", config.CONFIG.Network, id)

	return nil
}
