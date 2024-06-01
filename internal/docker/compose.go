package docker

import (
	"log"

	"github.com/compose-spec/compose-go/v2/cli"
	"github.com/compose-spec/compose-go/v2/types"
	"github.com/docker/compose/v2/pkg/api"
)

func (d *Api) Exec() {
	err := d.composeClient.Up(d.ctx, d.getDockerCompose(), api.UpOptions{
		Create: api.CreateOptions{
			RemoveOrphans: true,
			Recreate:      api.RecreateForce,
		},
	})
	if err != nil {
		log.Fatal(err)

	}
}

func (d *Api) getDockerCompose() *types.Project {
	composeFilePath := `G:\code\golang\sda\default_configs\compose_files\checkmk\docker-compose.yml`

	options, err := cli.NewProjectOptions(
		[]string{composeFilePath},
		cli.WithOsEnv,
		cli.WithDotEnv,
	)
	if err != nil {
		log.Fatal(err)
	}

	project, err := cli.ProjectFromOptions(d.ctx, options)
	if err != nil {
		log.Fatal(err)
	}

	return project
}
