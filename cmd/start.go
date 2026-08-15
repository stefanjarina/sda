package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/stefanjarina/sda/internal/docker"
	"github.com/stefanjarina/sda/internal/utils"
)

// startCmd represents the start command
var startCmd = &cobra.Command{
	Use:   "start [service]",
	Short: "Start a service",
	Long:  `Start a service`,
	Args:  bulkOrExactArgs,
	Run: func(cmd *cobra.Command, args []string) {
		sel, yes := readBulkFlags(cmd)
		client := docker.New()

		if sel.count() > 0 {
			runBulk(client, sel, bulkSpec{
				verb:    "start",
				started: "Started",
				empty:   "No services to start",
				prompt:  "Start",
				yes:     yes,
				action:  func(s docker.ServiceInfo) error { return client.Start(s.Name) },
			})
			return
		}

		name := args[0]
		service, err := lookupConfiguredService(name)
		if err != nil {
			utils.ErrorAndExit(err.Error())
		}
		if service.IsComposeService() {
			if err := client.ComposeStart(*service); err != nil {
				utils.ErrorAndExit(fmt.Sprintf("Failed to start compose service '%s': %v", name, err))
			}
			utils.Result(fmt.Sprintf("Started service '%s'", name))
			return
		}

		exists, err := client.Exists(name)
		if err != nil {
			utils.ErrorAndExit(fmt.Sprintf("Failed to check if service '%s' exists: %v", name, err))
		}
		if !exists {
			utils.ErrorAndExit(fmt.Sprintf("Service '%s' not found", name))
		}
		if err := client.Start(name); err != nil {
			utils.ErrorAndExit(fmt.Sprintf("Failed to start service '%s': %v", name, err))
		}
		utils.Result(fmt.Sprintf("Started service '%s'", name))
	},
}

func init() {
	rootCmd.AddCommand(startCmd)
	addBulkFlags(startCmd, "Start")
}
