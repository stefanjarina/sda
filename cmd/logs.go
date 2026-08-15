package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/stefanjarina/sda/internal/docker"
	"github.com/stefanjarina/sda/internal/utils"
)

var logsCmd = &cobra.Command{
	Use:   "logs [service]",
	Short: "Show service logs",
	Long:  `Show service logs`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]

		follow, _ := cmd.Flags().GetBool("follow")

		// Check if it's a compose service
		service, err := lookupConfiguredService(name)
		if err != nil {
			utils.ErrorAndExit(err.Error())
		}
		if service.IsComposeService() {
			// Handle as compose service
			client := docker.New()
			if err := client.ComposeLogs(*service, follow); err != nil {
				utils.ErrorAndExit(fmt.Sprintf("Failed to get logs: %v", err))
			}
			return
		}

		// Handle as Docker service
		tail, _ := cmd.Flags().GetInt("tail")
		timestamps, _ := cmd.Flags().GetBool("timestamps")

		client := docker.New()
		exists, err := client.Exists(name)
		if err != nil {
			utils.ErrorAndExit(fmt.Sprintf("Failed to check if service '%s' exists: %v", name, err))
		}
		if !exists {
			utils.ErrorAndExit(fmt.Sprintf("Service '%s' not found", name))
		}

		logOptions := docker.LogsOptions{
			Follow:     follow,
			Tail:       tail,
			Timestamps: timestamps,
		}

		if err := client.Logs(name, logOptions); err != nil {
			utils.ErrorAndExit(fmt.Sprintf("Failed to get logs: %v", err))
		}
	},
}

func init() {
	rootCmd.AddCommand(logsCmd)

	logsCmd.Flags().BoolP("follow", "f", false, "Follow log output")
	logsCmd.Flags().Int("tail", 100, "Number of lines to show from the end")
	logsCmd.Flags().Bool("timestamps", false, "Show timestamps")
}
