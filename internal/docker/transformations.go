package docker

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"

	"github.com/stefanjarina/sda/internal/config"
)

func expandVolumeSource(source, containerName string) (string, error) {
	return replacePlaceholder(source, map[string]string{"NAME": containerName})
}

func (d *Api) GetNamedVolumesForService(service *config.Service) ([]string, error) {
	if service == nil {
		return nil, nil
	}
	var volumes []string
	for _, v := range service.Docker.Volumes {
		if !v.IsNamed {
			continue
		}
		name, err := expandVolumeSource(v.Source, d.containerName(service.Name))
		if err != nil {
			return nil, fmt.Errorf("service %q volume %q: %w", service.Name, v.Source, err)
		}
		volumes = append(volumes, name)
	}
	return volumes, nil
}

func (d *Api) getNameFromContainerName(containerName string) string {
	return strings.TrimPrefix(containerName, d.cfg.Prefix+"-")
}

func getVersionFromImageName(imageName string) string {
	lastColon := strings.LastIndex(imageName, ":")
	lastSlash := strings.LastIndex(imageName, "/")
	if lastColon > lastSlash {
		tag := imageName[lastColon+1:]
		if tag != "" {
			return tag
		}
	}
	return "latest"
}

// parsePorts converts docker CLI's Ports column (e.g.
// "0.0.0.0:5432->5432/tcp, [::]:5432->5432/tcp") into ["5432:5432"],
// deduplicating the IPv4/IPv6 rows Docker reports for the same published
// port and skipping ports that are exposed but not published (no "->").
func parsePorts(portsStr string) []string {
	var ports []string
	seen := make(map[string]bool)

	for _, entry := range strings.Split(portsStr, ", ") {
		entry = strings.TrimSpace(entry)
		arrowIdx := strings.Index(entry, "->")
		if entry == "" || arrowIdx == -1 {
			continue
		}

		hostPart := entry[:arrowIdx]
		hostPort := hostPart[strings.LastIndex(hostPart, ":")+1:]

		containerPart := entry[arrowIdx+2:]
		if slashIdx := strings.Index(containerPart, "/"); slashIdx != -1 {
			containerPart = containerPart[:slashIdx]
		}

		mapping := fmt.Sprintf("%s:%s", hostPort, containerPart)
		if !seen[mapping] {
			seen[mapping] = true
			ports = append(ports, mapping)
		}
	}

	return ports
}

func replacePlaceholder(text string, obj any) (string, error) {
	templ, err := template.New("template").Option("missingkey=error").Parse(text)
	if err != nil {
		return "", fmt.Errorf("invalid template %q: %w", text, err)
	}
	var buf bytes.Buffer
	if err := templ.Execute(&buf, obj); err != nil {
		return "", fmt.Errorf("failed to render %q: %w", text, err)
	}
	return buf.String(), nil
}

func replacePassword(text string, service *config.Service, defaultPassword string) (string, error) {
	var password string
	if service.CustomPassword != "" {
		password = service.CustomPassword
	} else {
		password = defaultPassword
	}

	return replacePlaceholder(text, map[string]string{"PASSWORD": password})
}

// ValidateServiceTemplates checks env vars, volume sources, and the CLI
// connect command for parse errors and missing template keys.
func ValidateServiceTemplates(service *config.Service) error {
	if service == nil {
		return nil
	}
	passwordKeys := map[string]string{"PASSWORD": "x"}
	nameKeys := map[string]string{"NAME": "x"}

	for i, env := range service.Docker.EnvVars {
		if env == "" {
			continue
		}
		if _, err := replacePlaceholder(env, passwordKeys); err != nil {
			return fmt.Errorf("service %q: envVars[%d]: %w", service.Name, i, err)
		}
	}
	for i, v := range service.Docker.Volumes {
		if v.Source == "" {
			continue
		}
		if _, err := replacePlaceholder(v.Source, nameKeys); err != nil {
			return fmt.Errorf("service %q: volumes[%d].source: %w", service.Name, i, err)
		}
	}
	if service.CliConnectCommand != "" {
		if _, err := replacePlaceholder(service.CliConnectCommand, passwordKeys); err != nil {
			return fmt.Errorf("service %q: cliConnectCommand: %w", service.Name, err)
		}
	}
	return nil
}
