package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
	"sda/internal/config"
	"sda/internal/docker"
	"sda/internal/utils"
)

// newCmd represents the new command
var newCmd = &cobra.Command{
	Use:   "new [service]",
	Short: "Create new service",
	Long:  `Create new service`,
	Args:  cobra.ExactArgs(1),

	Run: func(cmd *cobra.Command, args []string) {
		serviceName := args[0]

		if !config.CONFIG.ServiceExists(serviceName) {
			utils.ErrorAndExit(fmt.Sprintf("Service '%s' does not exist\n", serviceName))
		}

		service := config.CONFIG.GetServiceByName(serviceName)

		networkName, _ := cmd.Flags().GetString("network")
		password, _ := cmd.Flags().GetString("password")
		version, _ := cmd.Flags().GetString("version")

		if password != "" {
			config.CONFIG.UpdatePassword(password)
			fmt.Printf("Creating '%s' in Docker container using custom password\n", service.OutputName)
		} else {
			fmt.Printf("Creating '%s' in Docker container using the default password '%s'\n", service.OutputName, config.CONFIG.Password)
			fmt.Printf("For custom password run again with: 'sda new %s -p <PASSWORD>'\n", serviceName)
			fmt.Println("Password must be strong, otherwise docker fails to create container")
		}

		answer := utils.Confirm("Are you sure you want to proceed? (Y/n): ")
		if !answer {
			os.Exit(0)
		}

		if networkName != "" {
			config.CONFIG.UpdateNetwork(networkName)
		}

		if version != "" {
			config.CONFIG.UpdateVersion(serviceName, version)
		}

		cli := docker.New()

		if !cli.CheckNetwork() {
			confirmedNetworkCreation := utils.Confirm(fmt.Sprintf("Network '%s' does not exist yet. Create it? (Y/n): ", networkName))
			if !confirmedNetworkCreation {
				utils.ErrorAndExit("Aborting, network must exist to create service")
			}
			_ = cli.CreateNetwork()
		}

		if err := cli.Create(serviceName); err != nil {
			utils.ErrorAndExit(fmt.Sprintf("Error creating container: %v", err))
		}

		fmt.Printf("Service '%s' created\n", service.OutputName)

		if err := cli.Start(serviceName); err != nil {
			utils.ErrorAndExit(fmt.Sprintf("Error starting container: %v", err))
		}

		fmt.Printf("Service '%s' started\n", service.OutputName)
	},
}

func init() {
	rootCmd.AddCommand(newCmd)

	newCmd.Flags().StringP("network", "n", "", "Network name")
	newCmd.Flags().StringP("password", "p", "", "Password")
	newCmd.Flags().String("version", "", "Version")
}
