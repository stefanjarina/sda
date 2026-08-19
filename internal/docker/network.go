package docker

import (
	"github.com/stefanjarina/sda/internal/utils"
)

func (d *Api) CheckNetwork(name string) bool {
	_, err := d.capture("network", "inspect", name)
	return err == nil
}

func (d *Api) CreateNetwork(name string) error {
	id, err := d.capture("network", "create", name)
	if err != nil {
		return err
	}

	utils.Progress("Created network '%s' with ID: %s\n", name, id)

	return nil
}
