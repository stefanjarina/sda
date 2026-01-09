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

		// Check if it's a compose service
		service := config.CONFIG.GetServiceByName(name)
		if service != nil && service.IsComposeService() {
			utils.Error(fmt.Sprintf("Service '%s' is a compose service", name))
			utils.ErrorAndExit("Connect command is not yet supported for compose services")
		}

		client := docker.New()
		if client.Exists(name) {
			password, _ := cmd.Flags().GetString("password")
			web, _ := cmd.Flags().GetBool("web")

			if err := client.Connect(name, password, web); err != nil {
				utils.Error(fmt.Sprintf("Failed to connect to service '%s': %v", name, err))
				utils.ErrorAndExit("")
			}
		} else {
			utils.Error(fmt.Sprintf("Service '%s' not found", name))
			utils.ErrorAndExit("")
		}
	},
}

func init() {
	rootCmd.AddCommand(connectCmd)

	connectCmd.Flags().BoolP("web", "w", false, "Open web browser instead of CLI")
	connectCmd.Flags().StringP("password", "p", "", "Password to use")
}
