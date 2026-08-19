package docker

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/stefanjarina/sda/internal/config"
	"github.com/stefanjarina/sda/internal/utils"
)

// psEntry is the subset of `docker ps --format {{json .}}` fields sda needs.
type psEntry struct {
	ID     string `json:"ID"`
	Names  string `json:"Names"`
	Image  string `json:"Image"`
	Ports  string `json:"Ports"`
	Status string `json:"Status"`
	State  string `json:"State"`
}

func (d *Api) ListAvailable() []ServiceInfo {
	var services []ServiceInfo

	for _, s := range d.cfg.Services {
		services = append(services, ServiceInfo{
			Name:          s.Name,
			ContainerName: d.containerName(s.Name),
			Image:         s.Docker.ImageName,
			Version:       s.Version,
			Ports:         []string{},
		})
	}

	return services
}

// psJSON runs `docker ps` narrowed by the given filter expressions (each
// passed as its own --filter) and decodes its newline-delimited JSON output.
func (d *Api) psJSON(filters ...string) ([]psEntry, error) {
	args := []string{"ps", "--all", "--no-trunc", "--format", "{{json .}}"}
	for _, f := range filters {
		args = append(args, "--filter", f)
	}

	out, err := d.capture(args...)
	if err != nil {
		return nil, err
	}
	if out == "" {
		return nil, nil
	}

	var entries []psEntry
	for _, line := range strings.Split(out, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		var e psEntry
		if err := json.Unmarshal([]byte(line), &e); err != nil {
			return nil, fmt.Errorf("failed to parse docker ps output: %w", err)
		}
		entries = append(entries, e)
	}

	return entries, nil
}

func (d *Api) toServiceInfo(e psEntry) ServiceInfo {
	statusIcon := "○"
	switch e.State {
	case "running":
		statusIcon = "●"
	case "exited":
		statusIcon = "✗"
	}

	return ServiceInfo{
		Name:          d.getNameFromContainerName(e.Names),
		ContainerName: e.Names,
		ID:            e.ID,
		Image:         e.Image,
		Version:       getVersionFromImageName(e.Image),
		Ports:         parsePorts(e.Ports),
		Status:        e.Status,
		StatusIcon:    statusIcon,
	}
}

// list returns containers managed by sda (name-prefixed), optionally
// narrowed to a status ("running", "exited", or "" for all).
func (d *Api) ownedNameFilter() string {
	return "name=^" + regexp.QuoteMeta(d.cfg.Prefix+"-")
}

func (d *Api) isOwnedContainer(name string) bool {
	name = strings.TrimPrefix(name, "/")
	return strings.HasPrefix(name, d.cfg.Prefix+"-")
}

func (d *Api) list(status string) ([]ServiceInfo, error) {
	filters := []string{d.ownedNameFilter()}
	if status != "" {
		filters = append(filters, "status="+status)
	}

	entries, err := d.psJSON(filters...)
	if err != nil {
		return nil, err
	}

	var services []ServiceInfo
	for _, e := range entries {
		if !d.isOwnedContainer(e.Names) {
			continue
		}
		services = append(services, d.toServiceInfo(e))
	}

	return services, nil
}

func (d *Api) ListCreated() ([]ServiceInfo, error) {
	return d.list("")
}

func (d *Api) ListRunning() ([]ServiceInfo, error) {
	return d.list("running")
}

func (d *Api) ListStopped() ([]ServiceInfo, error) {
	return d.list("exited")
}

// findContainer looks up a service's container by its exact container name.
// `docker ps --filter name=` matches substrings, so "sda-postgres" would also
// match "sda-postgres-replica" - the exact-name check guards against that.
func (d *Api) findContainer(name string) (ServiceInfo, bool, error) {
	name = d.containerName(name)

	entries, err := d.psJSON("name=" + name)
	if err != nil {
		return ServiceInfo{}, false, err
	}

	for _, e := range entries {
		if e.Names == name {
			return d.toServiceInfo(e), true, nil
		}
	}

	return ServiceInfo{}, false, nil
}

func (d *Api) GetInfo(name string) (ServiceInfo, error) {
	info, found, err := d.findContainer(name)
	if err != nil {
		return ServiceInfo{}, err
	}
	if !found {
		return ServiceInfo{}, fmt.Errorf("no container found for service %q", name)
	}

	return info, nil
}

func (d *Api) Exists(name string) (bool, error) {
	_, found, err := d.findContainer(name)
	return found, err
}

func (d *Api) Start(name string) error {
	_, err := d.capture("start", d.containerName(name))
	return err
}

func (d *Api) Stop(name string) error {
	_, err := d.capture("stop", d.containerName(name))
	return err
}

func (d *Api) Remove(name string, removeVolumes bool) error {
	args := []string{"rm", "--force"}
	if removeVolumes {
		args = append(args, "--volumes")
	}
	args = append(args, d.containerName(name))

	_, err := d.capture(args...)
	return err
}

func (d *Api) RemoveVolumes(names []string) error {
	var errs []error
	for _, name := range names {
		if _, err := d.capture("volume", "rm", "--force", name); err != nil {
			errs = append(errs, fmt.Errorf("volume %q: %w", name, err))
		}
	}
	return errors.Join(errs...)
}

func (d *Api) Connect(name string, customPassword string, web bool) error {
	service := d.cfg.GetServiceByName(name)
	if service == nil {
		return fmt.Errorf("service %q is not in the list of available services", name)
	}
	if web {
		if !service.HasWebConnect || service.WebConnectUrl == "" {
			return fmt.Errorf("service %q does not support web connect", name)
		}
		return handleWebConnect(service)
	}
	if !service.HasCliConnect {
		return fmt.Errorf("service %q does not support CLI connect", name)
	}
	return d.handleCliConnect(service, customPassword, name)
}

// Logs streams a container's logs directly to sda's own stdout/stderr.
func (d *Api) Logs(name string, opts LogsOptions) error {
	args := []string{"logs", "--tail", fmt.Sprintf("%d", opts.Tail)}
	if opts.Follow {
		args = append(args, "--follow")
	}
	if opts.Timestamps {
		args = append(args, "--timestamps")
	}
	args = append(args, d.containerName(name))

	return d.run(args...)
}

func (d *Api) containerName(name string) string {
	return fmt.Sprintf("%s-%s", d.cfg.Prefix, name)
}

func handleWebConnect(service *config.Service) error {
	return utils.OpenURL(service.WebConnectUrl)
}

func (d *Api) handleCliConnect(service *config.Service, customPassword, name string) error {
	tokens := splitArgs(service.CliConnectCommand)

	if service.HasPassword {
		passwordToUse := d.cfg.Password
		if customPassword != "" {
			passwordToUse = customPassword
		}
		for i, tok := range tokens {
			replaced, err := replacePassword(tok, service, passwordToUse)
			if err != nil {
				return err
			}
			tokens[i] = replaced
		}
	}

	args := append([]string{"exec", "-it", d.containerName(name)}, tokens...)
	return d.runInteractive(args...)
}
