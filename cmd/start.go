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
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]
		client := docker.New()
		if client.Exists(name) {
			err := client.Start(name)
			if err != nil {
				utils.Error(fmt.Sprintf("Failed to start service '%s': %v", name, err))
				utils.ErrorAndExit("")
			}

			fmt.Printf("Started service '%s'\n", name)
		} else {
			utils.Error(fmt.Sprintf("Service '%s' not found", name))
			utils.ErrorAndExit("")
		}
	},
}

func init() {
	rootCmd.AddCommand(startCmd)
}
