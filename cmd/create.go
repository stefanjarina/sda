package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/stefanjarina/sda/internal/config"
	"github.com/stefanjarina/sda/internal/docker"
	"github.com/stefanjarina/sda/internal/utils"
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
			utils.ErrorAndExit(fmt.Sprintf("Service '%s' is not in the list of available services", serviceName))
		}

		cli := docker.New()

		if cli.Exists(serviceName) {
			utils.ErrorAndExit(fmt.Sprintf("Service '%s' already exists", serviceName))
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
				fmt.Printf("Creating '%s' with custom password\n", service.OutputName)
			} else {
				fmt.Printf("Creating '%s' with default password\n", service.OutputName)
				fmt.Printf("For custom password run: 'sda create %s -p <PASSWORD>'\n", serviceName)
				fmt.Println("Password must be strong, otherwise Docker fails to create container")
			}
		} else {
			fmt.Printf("Creating '%s'\n", service.OutputName)
		}

		if !yes {
			answer := utils.Confirm("Proceed? (Y/n): ")
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
				confirmedNetworkCreation := utils.Confirm(fmt.Sprintf("Network '%s' does not exist. Create it? (Y/n): ", config.CONFIG.Network))
				if !confirmedNetworkCreation {
					utils.ErrorAndExit("Aborting: network must exist to create service")
				}
			}
			_ = cli.CreateNetwork()
		}

		if err := cli.Create(serviceName); err != nil {
			utils.Error(fmt.Sprintf("Failed to create container: %v", err))
			utils.ErrorAndExit("")
		}

		fmt.Printf("Created service '%s'\n", service.OutputName)

		if noStart {
			os.Exit(0)
		}

		if err := cli.Start(serviceName); err != nil {
			utils.Error(fmt.Sprintf("Failed to start container: %v", err))
			utils.ErrorAndExit("")
		}

		fmt.Printf("Started service '%s'\n", service.OutputName)
	},
}

func init() {
	rootCmd.AddCommand(createCmd)

	createCmd.Flags().StringP("network", "n", "", "Network name")
	createCmd.Flags().StringP("password", "p", "", "Password")
	createCmd.Flags().String("version", "", "Version")
	createCmd.Flags().Bool("no-start", false, "Do not start container after creation")
}
