package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stefanjarina/sda/internal/config"
	"github.com/stefanjarina/sda/internal/docker"
	"github.com/stefanjarina/sda/internal/utils"
)

// showCmd represents the show command
var showCmd = &cobra.Command{
	Use:   "show [service]",
	Short: "Show service information",
	Long:  `Show service information`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]

		// Check if it's a compose service
		service := config.CONFIG.GetServiceByName(name)
		if service != nil && service.IsComposeService() {
			utils.Error(fmt.Sprintf("Service '%s' is a compose service", name))
			utils.ErrorAndExit("Show command is not supported for compose services")
		}

		client := docker.New()
		if client.Exists(name) {
			serviceInfo := client.GetInfo(name)

			if viper.GetBool("json") {
				utils.Message(serviceInfo)
				return
			} else {
				fmt.Printf("Name: %s\n", serviceInfo.Name)
				fmt.Printf("Status: %s\n", serviceInfo.Status)
				fmt.Printf("Image: %s\n", serviceInfo.Image)
				fmt.Printf("Ports: %v\n", serviceInfo.Ports)
				fmt.Printf("ID: %s\n", serviceInfo.ID)
				fmt.Printf("Container Name: %s\n", serviceInfo.ContainerName)
			}

		} else {
			utils.Error(fmt.Sprintf("Service '%s' not found", name))
			utils.ErrorAndExit("")
		}
	},
}

func init() {
	rootCmd.AddCommand(showCmd)
}
