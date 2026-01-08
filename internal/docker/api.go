package docker

import (
	"context"

	"github.com/docker/cli/cli/command"
	"github.com/docker/cli/cli/flags"
	"github.com/docker/compose/v5/pkg/api"
	"github.com/docker/compose/v5/pkg/compose"
	"github.com/docker/docker/client"
)

type Api struct {
	client        *client.Client
	composeClient api.Compose
	ctx           context.Context
}

func New() *Api {
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}
	defer cli.Close()

	composeCli, err := initializeCompose()

	return &Api{
		client:        cli,
		ctx:           ctx,
		composeClient: composeCli,
	}
}

func initializeCompose() (api.Compose, error) {
	dockerCli, _ := command.NewDockerCli()

	dockerContext := "default"

	myOpts := &flags.ClientOptions{Context: dockerContext, LogLevel: "error"}
	_ = dockerCli.Initialize(myOpts)

	compose.NewComposeService(dockerCli)

	return compose.NewComposeService(dockerCli)
}
