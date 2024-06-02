package cmd

import (
	"fmt"
	"github.com/stefanjarina/sda/internal/config"
	"github.com/stefanjarina/sda/internal/docker"
	"github.com/stefanjarina/sda/internal/utils"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

var removeCmd = &cobra.Command{
	Use:   "remove [service]",
	Short: "Remove a service",
	Long:  `Remove a service`,
	Run: func(cmd *cobra.Command, args []string) {
		removeVolumes, _ := cmd.Flags().GetBool("volumes")
		yes, _ := cmd.Flags().GetBool("yes")

		name := args[0]
		client := docker.New()

		if client.Exists(name) {
			if !yes {
				confirmationMessage := fmt.Sprintf("Are you sure you want to remove '%s'? (Y/n): ", name)
				if removeVolumes {
					confirmationMessage = fmt.Sprintf("Are you sure you want to remove '%s' and all associated volumes? (Y/n): ", name)
				}

				confirmedRemove := utils.Confirm(confirmationMessage)
				if !confirmedRemove {
					os.Exit(0)
				}
			}

			fmt.Printf("Removing service '%s'...\n", name)
			err := client.Remove(name)
			if err != nil {
				utils.ErrorAndExit(fmt.Sprintf("Failed to remove service '%s': %v", name, err))
			}

			if removeVolumes {
				service := config.CONFIG.GetServiceByName(name)

				volumes := docker.GetNamedVolumesForService(service)

				if len(volumes) == 0 {
					os.Exit(0)
				}

				var confirmedVolumeRemove bool
				if !yes {
					confirmedVolumeRemove = utils.Confirm(fmt.Sprintf("These volumes will be removed: '%s' Proceed? (Y/n): ", strings.Join(volumes, ", ")))
				} else {
					confirmedVolumeRemove = true
				}
				if confirmedVolumeRemove {
					client.RemoveVolumes(volumes)
				}
			}

		} else {
			utils.ErrorAndExit(fmt.Sprintf("Service '%s' not found\n", name))
		}
	},
}

func init() {
	rootCmd.AddCommand(removeCmd)

	removeCmd.Flags().Bool("volumes", false, "Remove also volumes")
}
