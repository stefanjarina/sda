package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
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
