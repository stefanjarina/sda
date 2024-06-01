package cmd

import (
	"fmt"
	"github.com/stefanjarina/sda/internal/docker"
	"github.com/stefanjarina/sda/internal/utils"

	"github.com/spf13/cobra"
)

// stopCmd represents the stop command
var stopCmd = &cobra.Command{
	Use:   "stop [service]",
	Short: "Stop a service",
	Long:  `Stop a service`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]
		client := docker.New()
		if client.Exists(name) {
			err := client.Stop(name)
			if err != nil {
				utils.ErrorAndExit(fmt.Sprintf("Failed to stop service '%s': %v", name, err))
			}

			fmt.Printf("Service '%s' stopped\n", name)
		} else {
			utils.ErrorAndExit(fmt.Sprintf("Service '%s' not found\n", name))
		}
	},
}

func init() {
	rootCmd.AddCommand(stopCmd)
}
