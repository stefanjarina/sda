package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stefanjarina/sda/internal/config"
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
