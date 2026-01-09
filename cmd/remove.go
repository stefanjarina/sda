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

var removeCmd = &cobra.Command{
	Use:   "remove [service]",
	Short: "Remove a service",
	Long:  `Remove a service`,
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
		removeVolumes, _ := cmd.Flags().GetBool("volumes")
		yes, _ := cmd.Flags().GetBool("yes")
		all, _ := cmd.Flags().GetBool("all")
		running, _ := cmd.Flags().GetBool("running")
		stopped, _ := cmd.Flags().GetBool("stopped")

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
				fmt.Println("No services to remove")
				return
			}

			// Confirmation prompt
			if !yes {
				serviceNames := make([]string, len(services))
				for i, s := range services {
					serviceNames[i] = s.Name
				}
				confirmationMessage := fmt.Sprintf("Remove %s (%s)? (Y/n): ", actionDesc, strings.Join(serviceNames, ", "))
				if removeVolumes {
					confirmationMessage = fmt.Sprintf("Remove %s (%s) and all volumes? (Y/n): ", actionDesc, strings.Join(serviceNames, ", "))
				}
				confirmed := utils.Confirm(confirmationMessage)
				if !confirmed {
					os.Exit(0)
				}
			}

			// Execute bulk remove
			var failed []string
			var allVolumes []string

			for _, s := range services {
				err := client.Remove(s.Name)
				if err != nil {
					utils.Error(fmt.Sprintf("Failed to remove service '%s': %v", s.Name, err))
					failed = append(failed, s.Name)
				} else {
					fmt.Printf("Removed service '%s'\n", s.Name)

					// Collect volumes if needed
					if removeVolumes {
						service := config.CONFIG.GetServiceByName(s.Name)
						volumes := docker.GetNamedVolumesForService(service)
						allVolumes = append(allVolumes, volumes...)
					}
				}
			}

			// Handle volume removal
			if removeVolumes && len(allVolumes) > 0 {
				var confirmedVolumeRemove bool
				if !yes {
					confirmedVolumeRemove = utils.Confirm(fmt.Sprintf("Volumes to remove: %s. Proceed? (Y/n): ", strings.Join(allVolumes, ", ")))
				} else {
					confirmedVolumeRemove = true
				}
				if confirmedVolumeRemove {
					client.RemoveVolumes(allVolumes)
				}
			}

			if len(failed) > 0 {
				utils.Error(fmt.Sprintf("Failed to remove: %s", strings.Join(failed, ", ")))
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
			if !yes {
				confirmationMessage := fmt.Sprintf("Remove service '%s'? (Y/n): ", name)
				if removeVolumes {
					confirmationMessage = fmt.Sprintf("Remove service '%s' and all volumes? (Y/n): ", name)
				}

				confirmedRemove := utils.Confirm(confirmationMessage)
				if !confirmedRemove {
					os.Exit(0)
				}
			}

			fmt.Printf("Removing service '%s'...\n", name)
			if err := client.ComposeDown(*service, removeVolumes); err != nil {
				utils.Error(fmt.Sprintf("Failed to remove compose service '%s': %v", name, err))
				utils.ErrorAndExit("")
			}
			fmt.Printf("Removed service '%s'\n", name)
			return
		}

		// Handle as Docker service
		if client.Exists(name) {
			if !yes {
				confirmationMessage := fmt.Sprintf("Remove service '%s'? (Y/n): ", name)
				if removeVolumes {
					confirmationMessage = fmt.Sprintf("Remove service '%s' and all volumes? (Y/n): ", name)
				}

				confirmedRemove := utils.Confirm(confirmationMessage)
				if !confirmedRemove {
					os.Exit(0)
				}
			}

			fmt.Printf("Removing service '%s'...\n", name)
			err := client.Remove(name)
			if err != nil {
				utils.Error(fmt.Sprintf("Failed to remove service '%s': %v", name, err))
				utils.ErrorAndExit("")
			}

			if removeVolumes {
				service := config.CONFIG.GetServiceByName(name)

				volumes := docker.GetNamedVolumesForService(service)

				if len(volumes) == 0 {
					os.Exit(0)
				}

				var confirmedVolumeRemove bool
				if !yes {
					confirmedVolumeRemove = utils.Confirm(fmt.Sprintf("Volumes to remove: %s. Proceed? (Y/n): ", strings.Join(volumes, ", ")))
				} else {
					confirmedVolumeRemove = true
				}
				if confirmedVolumeRemove {
					client.RemoveVolumes(volumes)
				}
			}

		} else {
			utils.Error(fmt.Sprintf("Service '%s' not found", name))
			utils.ErrorAndExit("")
		}
	},
}

func init() {
	rootCmd.AddCommand(removeCmd)

	removeCmd.Flags().Bool("volumes", false, "Remove also volumes")
	removeCmd.Flags().Bool("all", false, "Remove all services")
	removeCmd.Flags().Bool("running", false, "Remove all running services")
	removeCmd.Flags().Bool("stopped", false, "Remove all stopped services")
}
