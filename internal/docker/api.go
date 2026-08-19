package docker

import (
	"fmt"
	"os/exec"

	"github.com/stefanjarina/sda/internal/config"
)

type Api struct {
	path string
	cfg  *config.Config
}

func New(cfg *config.Config) (*Api, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config is required")
	}
	path, err := exec.LookPath("docker")
	if err != nil {
		return nil, fmt.Errorf("Docker CLI not found in PATH. Install Docker: https://docs.docker.com/get-docker/")
	}
	return &Api{path: path, cfg: cfg}, nil
}
