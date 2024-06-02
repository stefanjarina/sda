package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/stefanjarina/sda/internal/config"
	"github.com/stefanjarina/sda/internal/docker"
	"github.com/stefanjarina/sda/internal/utils"
	"os"
)

// createCmd represents the new command
var createCmd = &cobra.Command{
	Use:   "create [service]",
	Short: "Create new service",
	Long:  `Create new service`,
	Args:  cobra.ExactArgs(1),

	Run: func(cmd *cobra.Command, args []string) {
		serviceName := args[0]

		if !config.CONFIG.ServiceExists(serviceName) {
			utils.ErrorAndExit(fmt.Sprintf("Service '%s' is not in a list of available services\n", serviceName))
		}

		cli := docker.New()

		if cli.Exists(serviceName) {
			utils.ErrorAndExit(fmt.Sprintf("Service '%s' already exists.\n", serviceName))
		}

		service := config.CONFIG.GetServiceByName(serviceName)

		networkName, _ := cmd.Flags().GetString("network")
		password, _ := cmd.Flags().GetString("password")
		version, _ := cmd.Flags().GetString("version")
		yes, _ := cmd.Flags().GetBool("yes")
		noStart, _ := cmd.Flags().GetBool("no-start")

		if service.HasPassword {
			if password != "" {
				config.CONFIG.UpdatePassword(password)
				fmt.Printf("Creating '%s' using custom password\n", service.OutputName)
			} else {
				fmt.Printf("Creating '%s' using the default password '%s'\n", service.OutputName, config.CONFIG.Password)
				fmt.Printf("For custom password run again with: 'sda new %s -p <PASSWORD>'\n", serviceName)
				fmt.Println("Password must be strong, otherwise docker fails to create container")
			}
		} else {
			fmt.Printf("Creating '%s'\n", service.OutputName)
		}

		if !yes {
			answer := utils.Confirm("Are you sure you want to proceed? (Y/n): ")
			if !answer {
				os.Exit(0)
			}
		}

		if networkName != "" {
			config.CONFIG.UpdateNetwork(networkName)
		}

		if version != "" {
			config.CONFIG.UpdateVersion(serviceName, version)
		}

		if !cli.CheckNetwork() {
			if !yes {
				confirmedNetworkCreation := utils.Confirm(fmt.Sprintf("Network '%s' does not exist yet. Create it? (Y/n): ", networkName))
				if !confirmedNetworkCreation {
					utils.ErrorAndExit("Aborting, network must exist to create service")
				}
			}
			_ = cli.CreateNetwork()
		}

		if err := cli.Create(serviceName); err != nil {
			utils.ErrorAndExit(fmt.Sprintf("Error creating container: %v", err))
		}

		fmt.Printf("Service '%s' created\n", service.OutputName)

		if noStart {
			os.Exit(0)
		}

		if err := cli.Start(serviceName); err != nil {
			utils.ErrorAndExit(fmt.Sprintf("Error starting container: %v", err))
		}

		fmt.Printf("Service '%s' started\n", service.OutputName)
	},
}

func init() {
	rootCmd.AddCommand(createCmd)

	createCmd.Flags().StringP("network", "n", "", "Network name")
	createCmd.Flags().StringP("password", "p", "", "Password")
	createCmd.Flags().String("version", "", "Version")
	createCmd.Flags().Bool("no-start", false, "Do not start container after creation")
}
