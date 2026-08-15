package docker

import (
	"reflect"
	"testing"

	"github.com/stefanjarina/sda/internal/config"
)

func TestBuildCreateArgs_Mssql(t *testing.T) {
	service := &config.Service{
		Name:           "mssql",
		HasPassword:    true,
		CustomPassword: "",
		Docker: config.Docker{
			ImageName: "mcr.microsoft.com/mssql/server",
			EnvVars:   []string{"ACCEPT_EULA=Y", "SA_PASSWORD={{.PASSWORD}}"},
			Volumes: []config.Volume{
				{Source: "{{.NAME}}-data", Target: "/var/opt/mssql", IsNamed: true},
			},
			PortMappings: []config.PortMapping{
				{Host: 1433, Container: 1433},
			},
			AdditionalDockerArguments: []string{"--ulimit nofile=262144:262144"},
		},
	}
	service.Version = "2022-latest"

	got, err := buildCreateArgs(service, "sda-mssql", "sda", "s3cr3t")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := []string{
		"create", "--name", "sda-mssql",
		"--network", "sda",
		"--env", "ACCEPT_EULA=Y",
		"--env", "SA_PASSWORD=s3cr3t",
		"--volume", "sda-mssql-data:/var/opt/mssql",
		"--publish", "1433:1433",
		"--ulimit", "nofile=262144:262144",
		"mcr.microsoft.com/mssql/server:2022-latest",
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("buildCreateArgs() =\n  %v\nwant\n  %v", got, want)
	}
}

func TestBuildCreateArgs_CustomAppCommands(t *testing.T) {
	service := &config.Service{
		Name: "redis-stack",
		Docker: config.Docker{
			ImageName:         "redis/redis-stack-server",
			PortMappings:      []config.PortMapping{{Host: 6379, Container: 6379}},
			CustomAppCommands: []string{"--appendonly yes"},
		},
	}
	service.Version = "latest"

	got, err := buildCreateArgs(service, "sda-redis-stack", "sda", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := []string{
		"create", "--name", "sda-redis-stack",
		"--network", "sda",
		"--publish", "6379:6379",
		"redis/redis-stack-server:latest",
		"--appendonly", "yes",
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("buildCreateArgs() =\n  %v\nwant\n  %v", got, want)
	}
}

func TestBuildCreateArgs_NoNetwork(t *testing.T) {
	service := &config.Service{
		Name:   "redis",
		Docker: config.Docker{ImageName: "redis"},
	}
	service.Version = "latest"

	got, err := buildCreateArgs(service, "sda-redis", "", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := []string{"create", "--name", "sda-redis", "redis:latest"}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("buildCreateArgs() =\n  %v\nwant\n  %v", got, want)
	}
}

func TestBuildCreateArgs_BadTemplate(t *testing.T) {
	service := &config.Service{
		Name: "postgres",
		Docker: config.Docker{
			ImageName: "postgres",
			EnvVars:   []string{"POSTGRES_PASSWORD={{.PASSWORD"},
		},
	}
	service.Version = "latest"

	_, err := buildCreateArgs(service, "sda-postgres", "sda", "secret")
	if err == nil {
		t.Fatal("expected error for malformed template")
	}
}
