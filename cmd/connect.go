package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/stefanjarina/sda/internal/config"
	"github.com/stefanjarina/sda/internal/docker"
	"github.com/stefanjarina/sda/internal/utils"
)

// connectCmd represents the connect command
var connectCmd = &cobra.Command{
	Use:   "connect [service]",
	Short: "Connects to a service",
	Long:  `Connects to a service via cli or opens a web browser if available`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]

		service := config.CONFIG.GetServiceByName(name)
		if service == nil {
			utils.ErrorAndExit(fmt.Sprintf("Service '%s' is not in the list of available services", name))
		}
		if service.IsComposeService() {
			utils.ErrorAndExit(fmt.Sprintf("Service '%s' is a compose service. Connect is not yet supported for compose services", name))
		}

		client := docker.New()
		exists, err := client.Exists(name)
		if err != nil {
			utils.ErrorAndExit(fmt.Sprintf("Failed to check if service '%s' exists: %v", name, err))
		}

		if exists {
			password, _ := cmd.Flags().GetString("password")
			web, _ := cmd.Flags().GetBool("web")

			if err := client.Connect(name, password, web); err != nil {
				utils.ErrorAndExit(fmt.Sprintf("Failed to connect to service '%s': %v", name, err))
			}
		} else {
			utils.ErrorAndExit(fmt.Sprintf("Service '%s' not found", name))
		}
	},
}

func init() {
	rootCmd.AddCommand(connectCmd)

	connectCmd.Flags().BoolP("web", "w", false, "Open web browser instead of CLI")
	connectCmd.Flags().StringP("password", "p", "", "Password to use")
}
