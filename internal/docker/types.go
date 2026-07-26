package docker

type ServiceInfo struct {
	Name          string
	ContainerName string
	ID            string
	Image         string
	Version       string
	Ports         []string
	Status        string
	StatusIcon    string
}

// LogsOptions controls `docker logs` output.
type LogsOptions struct {
	Follow     bool
	Timestamps bool
	Tail       int
}
