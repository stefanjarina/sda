package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/stefanjarina/sda/internal/config"
	"github.com/stefanjarina/sda/internal/docker"
	"github.com/stefanjarina/sda/internal/utils"

	"github.com/spf13/cobra"
)

var removeCmd = &cobra.Command{
	Use:   "remove [service]",
	Short: "Remove a service",
	Long:  `Remove a service`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		removeVolumes, _ := cmd.Flags().GetBool("volumes")
		yes, _ := cmd.Flags().GetBool("yes")

		name := args[0]
		client := docker.New()

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
}
