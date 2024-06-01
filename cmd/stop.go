package cmd

import (
	"github.com/cloudflare/cfssl/log"
	"sda/internal/config"
	"sda/internal/docker"

	"github.com/spf13/cobra"
)

// stopCmd represents the stop command
var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop a service",
	Long:  `Stop a service`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]
		if service := config.CONFIG.GetServiceByName(name); service != nil {
			client := docker.New()
			err := client.Stop(name)
			if err != nil {
				log.Errorf("Failed to start service %s: %v", name, err)
			}
		} else {
			log.Errorf("Service %s not found", name)
		}
	},
}

func init() {
	rootCmd.AddCommand(stopCmd)
}
