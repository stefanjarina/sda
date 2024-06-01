package docker

import (
	"bytes"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/go-units"
	"github.com/stefanjarina/sda/internal/config"
	"regexp"
	"strconv"
	"strings"
	"text/template"
)

func GetNamedVolumesForService(service *config.Service) []string {
	var volumes []string
	for _, v := range service.Docker.Volumes {
		if v.IsNamed {
			volumes = append(volumes, replacePlaceholder(v.Source, map[string]string{"NAME": service.Name}))
		}
	}
	return volumes
}

func getNameFromContainerName(containerName string) string {
	containerName = containerName[len(config.CONFIG.Prefix)+2:]
	return containerName
}

func getVersionFromImageName(imageName string) string {
	imageName = imageName[strings.LastIndex(imageName, ":")+1:]
	return imageName
}

func getPortsFromContainer(containerPorts []types.Port) []string {
	var ports []string
	for _, port := range containerPorts {
		ports = append(ports, fmt.Sprintf("%d:%d", port.PublicPort, port.PrivatePort))
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

func parseUlimits(args []string) []*units.Ulimit {
	var ulimits []*units.Ulimit
	for _, arg := range args {
		ulimit := parseUlimit(arg)
		if ulimit != nil {
			ulimits = append(ulimits, ulimit)
		}
	}
	return ulimits
}

func parseUlimit(text string) *units.Ulimit {
	var r = regexp.MustCompile(`--ulimit\s(?P<name>\w+)=(?P<soft>\d+):(?P<hard>\d+)`)
	var u *units.Ulimit
	match := r.FindStringSubmatch(text)
	matchNames := r.SubexpNames()
	if match != nil {
		name := getMatchByName("name", match, matchNames)
		soft, _ := strconv.ParseInt(getMatchByName("soft", match, matchNames), 10, 64)
		hard, _ := strconv.ParseInt(getMatchByName("hard", match, matchNames), 10, 64)
		u = &units.Ulimit{
			Name: name,
			Soft: soft,
			Hard: hard,
		}
	}
	return u
}

func getMatchByName(name string, match []string, matchNames []string) string {
	for i, matchName := range matchNames {
		if matchName == name {
			return match[i]
		}
	}
	return ""
}
