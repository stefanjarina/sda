package docker

import (
	"os/exec"

	"github.com/stefanjarina/sda/internal/utils"
)

type Api struct {
	path string
}

func New() *Api {
	path, err := exec.LookPath("docker")
	if err != nil {
		utils.ErrorAndExit("Docker CLI not found in PATH. Install Docker: https://docs.docker.com/get-docker/")
	}

	return &Api{
		path: path,
	}
}
