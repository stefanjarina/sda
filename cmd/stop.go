package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/stefanjarina/sda/internal/config"
	"github.com/stefanjarina/sda/internal/docker"
	"github.com/stefanjarina/sda/internal/utils"
)

// stopCmd represents the stop command
var stopCmd = &cobra.Command{
	Use:   "stop [service]",
	Short: "Stop a service",
	Long:  `Stop a service`,
	Args: func(cmd *cobra.Command, args []string) error {
		all, _ := cmd.Flags().GetBool("all")
		running, _ := cmd.Flags().GetBool("running")
		stopped, _ := cmd.Flags().GetBool("stopped")

		if all || running || stopped {
			if len(args) > 0 {
				return fmt.Errorf("cannot specify service name with bulk flags")
			}
			return nil
		}
		return cobra.ExactArgs(1)(cmd, args)
	},
	Run: func(cmd *cobra.Command, args []string) {
		all, _ := cmd.Flags().GetBool("all")
		running, _ := cmd.Flags().GetBool("running")
		stopped, _ := cmd.Flags().GetBool("stopped")
		yes, _ := cmd.Flags().GetBool("yes")

		// Validate mutual exclusivity
		flagCount := 0
		if all {
			flagCount++
		}
		if running {
			flagCount++
		}
		if stopped {
			flagCount++
		}
		if flagCount > 1 {
			utils.Error("Only one of --all, --running, or --stopped can be specified")
			utils.ErrorAndExit("")
		}

		client := docker.New()

		// Handle bulk operations
		if all || running || stopped {
			var services []docker.ServiceInfo
			var actionDesc string

			if all {
				services = client.ListCreated()
				actionDesc = "all services"
			} else if running {
				services = client.ListRunning()
				actionDesc = "all running services"
			} else if stopped {
				services = client.ListStopped()
				actionDesc = "all stopped services"
			}

			if len(services) == 0 {
				fmt.Println("No services to stop")
				return
			}

			// Confirmation prompt
			if !yes {
				serviceNames := make([]string, len(services))
				for i, s := range services {
					serviceNames[i] = s.Name
				}
				confirmed := utils.Confirm(fmt.Sprintf("Stop %s (%s)? (Y/n): ", actionDesc, strings.Join(serviceNames, ", ")))
				if !confirmed {
					os.Exit(0)
				}
			}

			// Execute bulk stop
			var failed []string
			for _, s := range services {
				err := client.Stop(s.Name)
				if err != nil {
					utils.Error(fmt.Sprintf("Failed to stop service '%s': %v", s.Name, err))
					failed = append(failed, s.Name)
				} else {
					fmt.Printf("Stopped service '%s'\n", s.Name)
				}
			}

			if len(failed) > 0 {
				utils.Error(fmt.Sprintf("Failed to stop: %s", strings.Join(failed, ", ")))
				utils.ErrorAndExit("")
			}
			return
		}

		// Handle single service
		name := args[0]

		// Check if it's a compose service
		service := config.CONFIG.GetServiceByName(name)
		if service != nil && service.IsComposeService() {
			// Handle as compose service
			if err := client.ComposeStop(*service); err != nil {
				utils.Error(fmt.Sprintf("Failed to stop compose service '%s': %v", name, err))
				utils.ErrorAndExit("")
			}
			fmt.Printf("Stopped service '%s'\n", name)
			return
		}

		// Handle as Docker service
		if client.Exists(name) {
			err := client.Stop(name)
			if err != nil {
				utils.Error(fmt.Sprintf("Failed to stop service '%s': %v", name, err))
				utils.ErrorAndExit("")
			}

			fmt.Printf("Stopped service '%s'\n", name)
		} else {
			utils.Error(fmt.Sprintf("Service '%s' not found", name))
			utils.ErrorAndExit("")
		}
	},
}

func init() {
	rootCmd.AddCommand(stopCmd)

	stopCmd.Flags().Bool("all", false, "Stop all services")
	stopCmd.Flags().Bool("running", false, "Stop all running services")
	stopCmd.Flags().Bool("stopped", false, "Stop all stopped services")
}
