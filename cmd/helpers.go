package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stefanjarina/sda/internal/config"
	"github.com/stefanjarina/sda/internal/docker"
	"github.com/stefanjarina/sda/internal/utils"
)

type listMode int

const (
	listRunning listMode = iota
	listStopped
	listCreated
	listAvailable
	listCompose
)

func selectListMode(available, created, running, stopped, compose bool) (listMode, error) {
	n := 0
	if available {
		n++
	}
	if created {
		n++
	}
	if running {
		n++
	}
	if stopped {
		n++
	}
	if compose {
		n++
	}
	if n > 1 {
		return 0, fmt.Errorf("only one of --available, --created, --running, --stopped, or --compose can be specified")
	}
	switch {
	case available:
		return listAvailable, nil
	case created:
		return listCreated, nil
	case stopped:
		return listStopped, nil
	case compose:
		return listCompose, nil
	default:
		return listRunning, nil
	}
}

func wantsJSON(cmd *cobra.Command) bool {
	if viper.GetBool("json") {
		return true
	}
	format, err := cmd.Flags().GetString("format")
	return err == nil && format == "json"
}

func requireRecreateForVolumes(recreate, volumes bool) error {
	if volumes && !recreate {
		return fmt.Errorf("--volumes flag requires --recreate flag")
	}
	return nil
}

func rejectNoOpBulk(verb string, running, stopped bool) error {
	if verb == "start" && running {
		return fmt.Errorf("cannot use --running with start; services are already running (use --stopped or --all)")
	}
	if verb == "stop" && stopped {
		return fmt.Errorf("cannot use --stopped with stop; services are already stopped (use --running or --all)")
	}
	return nil
}

func lookupConfiguredService(name string) (*config.Service, error) {
	svc := config.CONFIG.GetServiceByName(name)
	if svc == nil {
		return nil, fmt.Errorf("Service '%s' is not in the list of available services", name)
	}
	return svc, nil
}

type bulkSelector struct {
	all, running, stopped bool
}

func (s bulkSelector) count() int {
	n := 0
	if s.all {
		n++
	}
	if s.running {
		n++
	}
	if s.stopped {
		n++
	}
	return n
}

func (s bulkSelector) validate() error {
	if s.count() > 1 {
		return fmt.Errorf("Only one of --all, --running, or --stopped can be specified")
	}
	return nil
}

func (s bulkSelector) list(client *docker.Api) ([]docker.ServiceInfo, string, error) {
	switch {
	case s.all:
		svcs, err := client.ListCreated()
		return svcs, "all services", err
	case s.running:
		svcs, err := client.ListRunning()
		return svcs, "all running services", err
	case s.stopped:
		svcs, err := client.ListStopped()
		return svcs, "all stopped services", err
	default:
		return nil, "", fmt.Errorf("no bulk selector set")
	}
}

func readBulkFlags(cmd *cobra.Command) (bulkSelector, bool) {
	all, _ := cmd.Flags().GetBool("all")
	running, _ := cmd.Flags().GetBool("running")
	stopped, _ := cmd.Flags().GetBool("stopped")
	yes, _ := cmd.Flags().GetBool("yes")
	return bulkSelector{all: all, running: running, stopped: stopped}, yes
}

func addBulkFlags(cmd *cobra.Command, verb string) {
	cmd.Flags().Bool("all", false, fmt.Sprintf("%s all services", verb))
	cmd.Flags().Bool("running", false, fmt.Sprintf("%s all running services", verb))
	cmd.Flags().Bool("stopped", false, fmt.Sprintf("%s all stopped services", verb))
}

func bulkOrExactArgs(cmd *cobra.Command, args []string) error {
	sel, _ := readBulkFlags(cmd)
	if sel.count() > 0 {
		if len(args) > 0 {
			return fmt.Errorf("cannot specify service name with bulk flags")
		}
		return nil
	}
	return cobra.ExactArgs(1)(cmd, args)
}

func bulkPrompt(prompt, actionDesc string, names []string, volumes bool) string {
	joined := strings.Join(names, ", ")
	if volumes {
		return fmt.Sprintf("%s %s (%s) and all volumes? (Y/n): ", prompt, actionDesc, joined)
	}
	return fmt.Sprintf("%s %s (%s)? (Y/n): ", prompt, actionDesc, joined)
}

type bulkItem struct {
	Service string `json:"service"`
	Error   string `json:"error,omitempty"`
}

type bulkDocument struct {
	OK       []string   `json:"ok"`
	Failed   []bulkItem `json:"failed"`
	Warnings []string   `json:"warnings,omitempty"`
}

type bulkOutcome struct {
	verb    string
	started string
	ok      []string
	failed  []bulkItem
	warns   []string
}

func newBulkOutcome(verb, started string) *bulkOutcome {
	return &bulkOutcome{
		verb:    verb,
		started: started,
		ok:      []string{},
		failed:  []bulkItem{},
	}
}

func (o *bulkOutcome) record(name string, err error) {
	if err != nil {
		o.failed = append(o.failed, bulkItem{Service: name, Error: err.Error()})
		if !utils.JSONMode() {
			utils.Error(fmt.Sprintf("Failed to %s service '%s': %v", o.verb, name, err))
		}
		return
	}
	o.ok = append(o.ok, name)
	if !utils.JSONMode() {
		fmt.Printf("%s service '%s'\n", o.started, name)
	}
}

func (o *bulkOutcome) warn(msg string) {
	o.warns = append(o.warns, msg)
	if !utils.JSONMode() {
		utils.Error(msg)
	}
}

func (o *bulkOutcome) document() bulkDocument {
	return bulkDocument{OK: o.ok, Failed: o.failed, Warnings: o.warns}
}

func (o *bulkOutcome) finish() {
	if utils.JSONMode() {
		utils.JSON(o.document())
		if len(o.failed) > 0 {
			utils.ErrorAndExit("")
		}
		return
	}
	if len(o.failed) > 0 {
		names := make([]string, len(o.failed))
		for i, f := range o.failed {
			names[i] = f.Service
		}
		utils.ErrorAndExit(fmt.Sprintf("Failed to %s: %s", o.verb, strings.Join(names, ", ")))
	}
}

func emitEmptyBulk(msg string) {
	if utils.JSONMode() {
		utils.JSON(bulkDocument{OK: []string{}, Failed: []bulkItem{}})
		return
	}
	fmt.Println(msg)
}

type bulkSpec struct {
	verb     string
	started  string
	empty    string
	prompt   string
	yes      bool
	volumes  bool
	action   func(docker.ServiceInfo) error
	afterOK  func(docker.ServiceInfo, *bulkOutcome)
	afterAll func(*bulkOutcome)
}

func runBulk(client *docker.Api, sel bulkSelector, spec bulkSpec) {
	if err := sel.validate(); err != nil {
		utils.ErrorAndExit(err.Error())
	}
	if err := rejectNoOpBulk(spec.verb, sel.running, sel.stopped); err != nil {
		utils.ErrorAndExit(err.Error())
	}
	services, actionDesc, err := sel.list(client)
	if err != nil {
		utils.ErrorAndExit(fmt.Sprintf("Failed to list services: %v", err))
	}
	if len(services) == 0 {
		emitEmptyBulk(spec.empty)
		return
	}
	if !spec.yes {
		names := make([]string, len(services))
		for i, s := range services {
			names[i] = s.Name
		}
		if !utils.Confirm(bulkPrompt(spec.prompt, actionDesc, names, spec.volumes)) {
			utils.Cancelled()
		}
	}
	outcome := newBulkOutcome(spec.verb, spec.started)
	for _, s := range services {
		err := spec.action(s)
		outcome.record(s.Name, err)
		if err == nil && spec.afterOK != nil {
			spec.afterOK(s, outcome)
		}
	}
	if spec.afterAll != nil {
		spec.afterAll(outcome)
	}
	outcome.finish()
}
