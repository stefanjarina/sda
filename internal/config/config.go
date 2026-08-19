package config

var CONFIG Config

type Config struct {
	Network  string    `mapstructure:"defaultNetwork"`
	Password string    `mapstructure:"defaultPassword"`
	Prefix   string    `mapstructure:"prefix"`
	Services []Service `mapstructure:"services"`
}

type Service struct {
	Name              string `mapstructure:"name"`
	OutputName        string `mapstructure:"outputName"`
	Version           string `mapstructure:"defaultVersion"`
	HasPassword       bool   `mapstructure:"hasPassword"`
	CustomPassword    string `mapstructure:"customPassword"`
	Compose           string `mapstructure:"compose"`
	Docker            Docker `mapstructure:"docker"`
	HasCliConnect     bool   `mapstructure:"hasCliConnect"`
	CliConnectCommand string `mapstructure:"cliConnectCommand"`
	HasWebConnect     bool   `mapstructure:"hasWebConnect"`
	WebConnectUrl     string `mapstructure:"webConnectUrl"`
}

// IsComposeService returns true if the service is configured with a compose file
func (s *Service) IsComposeService() bool {
	return s.Compose != ""
}

type Docker struct {
	ImageName                 string        `mapstructure:"imageName"`
	PortMappings              []PortMapping `mapstructure:"portMappings"`
	IsPersistent              bool          `mapstructure:"isPersistent"`
	Volumes                   []Volume      `mapstructure:"volumes"`
	EnvVars                   []string      `mapstructure:"envVars"`
	AdditionalDockerArguments []string      `mapstructure:"additionalDockerArguments"`
	CustomAppCommands         []string      `mapstructure:"customAppCommands"`
}

type PortMapping struct {
	Host      int `mapstructure:"host"`
	Container int `mapstructure:"container"`
}

type Volume struct {
	Source  string `mapstructure:"source"`
	Target  string `mapstructure:"target"`
	IsNamed bool   `mapstructure:"isNamed"`
}

func (c *Config) GetServiceByName(name string) *Service {
	for i := range c.Services {
		if c.Services[i].Name == name {
			return &c.Services[i]
		}
	}

	return nil
}

func (c *Config) GetAllServiceNames() []string {
	var names []string
	for _, service := range c.Services {
		names = append(names, service.Name)
	}

	return names
}

func (c *Config) ServiceExists(name string) bool {
	for _, service := range c.Services {
		if service.Name == name {
			return true
		}
	}

	return false
}
