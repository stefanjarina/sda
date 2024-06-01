package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"sda/internal/docker"
	"sda/internal/utils"
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
				utils.ErrorAndExit(fmt.Sprintf("Failed to start service '%s': %v\n", name, err))
			}

			fmt.Printf("Service '%s' started\n", name)
		} else {
			utils.ErrorAndExit(fmt.Sprintf("Service '%s' not found\n", name))
		}
	},
}

func init() {
	rootCmd.AddCommand(startCmd)
}
