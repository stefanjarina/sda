package cmd

import (
	"github.com/cloudflare/cfssl/log"
	"github.com/spf13/cobra"
	"sda/internal/config"
	"sda/internal/docker"
)

// startCmd represents the start command
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start a service",
	Long:  `Start a service`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]
		if service := config.CONFIG.GetServiceByName(name); service != nil {
			client := docker.New()
			err := client.Start(name)
			if err != nil {
				log.Errorf("Failed to start service %s: %v", name, err)
			}
		} else {
			log.Errorf("Service %s not found", name)
		}
	},
}

func init() {
	rootCmd.AddCommand(startCmd)
}
