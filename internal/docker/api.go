package docker

import (
	"context"
	"fmt"

	"github.com/docker/docker/client"
	"github.com/stefanjarina/sda/internal/utils"
)

type Api struct {
	client *client.Client
	ctx    context.Context
}

func New() *Api {
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		utils.ErrorAndExit(fmt.Sprintf("Failed to create Docker client: %v", err))
	}

	return &Api{
		client: cli,
		ctx:    ctx,
	}
}
