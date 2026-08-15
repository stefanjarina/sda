package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/stefanjarina/sda/internal/docker"
	"github.com/stefanjarina/sda/internal/utils"
)

// stopCmd represents the stop command
var stopCmd = &cobra.Command{
	Use:   "stop [service]",
	Short: "Stop a service",
	Long:  `Stop a service`,
	Args:  bulkOrExactArgs,
	Run: func(cmd *cobra.Command, args []string) {
		sel, yes := readBulkFlags(cmd)
		client := docker.New()

		if sel.count() > 0 {
			runBulk(client, sel, bulkSpec{
				verb:    "stop",
				started: "Stopped",
				empty:   "No services to stop",
				prompt:  "Stop",
				yes:     yes,
				action:  func(s docker.ServiceInfo) error { return client.Stop(s.Name) },
			})
			return
		}

		name := args[0]
		service, err := lookupConfiguredService(name)
		if err != nil {
			utils.ErrorAndExit(err.Error())
		}
		if service.IsComposeService() {
			if err := client.ComposeStop(*service); err != nil {
				utils.ErrorAndExit(fmt.Sprintf("Failed to stop compose service '%s': %v", name, err))
			}
			utils.Result(fmt.Sprintf("Stopped service '%s'", name))
			return
		}

		exists, err := client.Exists(name)
		if err != nil {
			utils.ErrorAndExit(fmt.Sprintf("Failed to check if service '%s' exists: %v", name, err))
		}
		if !exists {
			utils.ErrorAndExit(fmt.Sprintf("Service '%s' not found", name))
		}
		if err := client.Stop(name); err != nil {
			utils.ErrorAndExit(fmt.Sprintf("Failed to stop service '%s': %v", name, err))
		}
		utils.Result(fmt.Sprintf("Stopped service '%s'", name))
	},
}

func init() {
	rootCmd.AddCommand(stopCmd)
	addBulkFlags(stopCmd, "Stop")
}
