package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
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

		service, err := lookupConfiguredService(name)
		if err != nil {
			utils.ErrorAndExit(err.Error())
		}
		if service.IsComposeService() {
			utils.ErrorAndExit(fmt.Sprintf("Service '%s' is a compose service. Show is not supported for compose services", name))
		}

		client := docker.New()
		exists, err := client.Exists(name)
		if err != nil {
			utils.ErrorAndExit(fmt.Sprintf("Failed to check if service '%s' exists: %v", name, err))
		}

		if exists {
			serviceInfo, err := client.GetInfo(name)
			if err != nil {
				utils.ErrorAndExit(fmt.Sprintf("Failed to get info for service '%s': %v", name, err))
			}

			if wantsJSON(cmd) {
				utils.JSON(serviceInfo)
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
			utils.ErrorAndExit(fmt.Sprintf("Service '%s' not found", name))
		}
	},
}

func init() {
	rootCmd.AddCommand(showCmd)

	showCmd.Flags().String("format", "table", "Output format: table, json")
}
