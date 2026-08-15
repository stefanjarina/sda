package docker

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"

	"github.com/stefanjarina/sda/internal/config"
)

func GetNamedVolumesForService(service *config.Service) []string {
	if service == nil {
		return nil
	}
	var volumes []string
	for _, v := range service.Docker.Volumes {
		if v.IsNamed {
			volumes = append(volumes, replacePlaceholder(v.Source, map[string]string{"NAME": service.Name}))
		}
	}
	return volumes
}

func getNameFromContainerName(containerName string) string {
	return strings.TrimPrefix(containerName, config.CONFIG.Prefix+"-")
}

func getVersionFromImageName(imageName string) string {
	imageName = imageName[strings.LastIndex(imageName, ":")+1:]
	return imageName
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

func replacePlaceholder(text string, obj any) string {
	var buf bytes.Buffer
	templ := template.Must(template.New("template").Parse(text))
	_ = templ.Execute(&buf, obj)
	return buf.String()
}

func replacePassword(text string, service *config.Service, defaultPassword string) string {
	var password string
	if service.CustomPassword != "" {
		password = service.CustomPassword
	} else {
		password = defaultPassword
	}

	result := replacePlaceholder(text, map[string]string{"PASSWORD": password})
	return result
}
