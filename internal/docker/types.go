package docker

type ServiceInfo struct {
	Name          string
	ContainerName string
	ID            string
	Image         string
	Version       string
	Ports         []string
	Status        string
}
