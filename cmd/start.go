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

// startCmd represents the start command
var startCmd = &cobra.Command{
	Use:   "start [service]",
	Short: "Start a service",
	Long:  `Start a service`,
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
				fmt.Println("No services to start")
				return
			}

			// Confirmation prompt
			if !yes {
				serviceNames := make([]string, len(services))
				for i, s := range services {
					serviceNames[i] = s.Name
				}
				confirmed := utils.Confirm(fmt.Sprintf("Start %s (%s)? (Y/n): ", actionDesc, strings.Join(serviceNames, ", ")))
				if !confirmed {
					os.Exit(0)
				}
			}

			// Execute bulk start
			var failed []string
			for _, s := range services {
				err := client.Start(s.Name)
				if err != nil {
					utils.Error(fmt.Sprintf("Failed to start service '%s': %v", s.Name, err))
					failed = append(failed, s.Name)
				} else {
					fmt.Printf("Started service '%s'\n", s.Name)
				}
			}

			if len(failed) > 0 {
				utils.Error(fmt.Sprintf("Failed to start: %s", strings.Join(failed, ", ")))
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
			if err := client.ComposeStart(*service); err != nil {
				utils.Error(fmt.Sprintf("Failed to start compose service '%s': %v", name, err))
				utils.ErrorAndExit("")
			}
			fmt.Printf("Started service '%s'\n", name)
			return
		}

		// Handle as Docker service
		if client.Exists(name) {
			err := client.Start(name)
			if err != nil {
				utils.Error(fmt.Sprintf("Failed to start service '%s': %v", name, err))
				utils.ErrorAndExit("")
			}

			fmt.Printf("Started service '%s'\n", name)
		} else {
			utils.Error(fmt.Sprintf("Service '%s' not found", name))
			utils.ErrorAndExit("")
		}
	},
}

func init() {
	rootCmd.AddCommand(startCmd)

	startCmd.Flags().Bool("all", false, "Start all services")
	startCmd.Flags().Bool("running", false, "Start all running services")
	startCmd.Flags().Bool("stopped", false, "Start all stopped services")
}
