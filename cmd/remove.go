package cmd

import (
	"fmt"
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
	Args:  bulkOrExactArgs,
	Run: func(cmd *cobra.Command, args []string) {
		removeVolumes, _ := cmd.Flags().GetBool("volumes")
		sel, yes := readBulkFlags(cmd)
		client := mustDockerClient()

		if sel.count() > 0 {
			var allVolumes []string
			runBulk(client, sel, bulkSpec{
				verb:    "remove",
				started: "Removed",
				empty:   "No services to remove",
				prompt:  "Remove",
				yes:     yes,
				volumes: removeVolumes,
				action:  func(s docker.ServiceInfo) error { return client.Remove(s.Name, removeVolumes) },
				afterOK: func(s docker.ServiceInfo, o *bulkOutcome) {
					if !removeVolumes {
						return
					}
					service := config.CONFIG.GetServiceByName(s.Name)
					if service == nil {
						return
					}
					volumes, err := client.GetNamedVolumesForService(service)
					if err != nil {
						o.warn(fmt.Sprintf("Failed to resolve volumes for '%s': %v", s.Name, err))
						return
					}
					allVolumes = append(allVolumes, volumes...)
				},
				afterAll: func(o *bulkOutcome) {
					if !removeVolumes || len(allVolumes) == 0 {
						return
					}
					confirmed := yes
					if !yes {
						confirmed = utils.Confirm(fmt.Sprintf("Volumes to remove: %s. Proceed? (Y/n): ", strings.Join(allVolumes, ", ")))
					}
					if confirmed {
						if err := client.RemoveVolumes(allVolumes); err != nil {
							o.warn(fmt.Sprintf("Failed to remove volumes: %v", err))
						}
					}
				},
			})
			return
		}

		name := args[0]

		service := config.CONFIG.GetServiceByName(name)
		if service != nil && service.IsComposeService() {
			if !yes {
				confirmationMessage := fmt.Sprintf("Remove service '%s'? (Y/n): ", name)
				if removeVolumes {
					confirmationMessage = fmt.Sprintf("Remove service '%s' and all volumes? (Y/n): ", name)
				}

				confirmedRemove := utils.Confirm(confirmationMessage)
				if !confirmedRemove {
					utils.Cancelled()
				}
			}

			utils.Progress("Removing service '%s'...\n", name)
			if err := client.ComposeDown(*service, removeVolumes); err != nil {
				utils.ErrorAndExit(fmt.Sprintf("Failed to remove compose service '%s': %v", name, err))
			}
			utils.Result(fmt.Sprintf("Removed service '%s'", name))
			return
		}

		exists, err := client.Exists(name)
		if err != nil {
			utils.ErrorAndExit(fmt.Sprintf("Failed to check if service '%s' exists: %v", name, err))
		}

		if !exists {
			utils.ErrorAndExit(fmt.Sprintf("Service '%s' not found", name))
		}

		if !yes {
			confirmationMessage := fmt.Sprintf("Remove service '%s'? (Y/n): ", name)
			if removeVolumes {
				confirmationMessage = fmt.Sprintf("Remove service '%s' and all volumes? (Y/n): ", name)
			}

			confirmedRemove := utils.Confirm(confirmationMessage)
			if !confirmedRemove {
				utils.Cancelled()
			}
		}

		utils.Progress("Removing service '%s'...\n", name)
		if err := client.Remove(name, removeVolumes); err != nil {
			utils.ErrorAndExit(fmt.Sprintf("Failed to remove service '%s': %v", name, err))
		}

		var volumeErr error
		if removeVolumes {
			service := config.CONFIG.GetServiceByName(name)
			volumes, err := client.GetNamedVolumesForService(service)
			if err != nil {
				utils.ErrorAndExit(fmt.Sprintf("Failed to resolve volumes: %v", err))
			}

			if len(volumes) == 0 {
				utils.Result(fmt.Sprintf("Removed service '%s'", name))
				return
			}

			confirmedVolumeRemove := yes
			if !yes {
				confirmedVolumeRemove = utils.Confirm(fmt.Sprintf("Volumes to remove: %s. Proceed? (Y/n): ", strings.Join(volumes, ", ")))
			}
			if confirmedVolumeRemove {
				volumeErr = client.RemoveVolumes(volumes)
			}
		}

		if volumeErr != nil {
			if utils.JSONMode() {
				utils.JSON(map[string]any{
					"ok":       true,
					"message":  fmt.Sprintf("Removed service '%s'", name),
					"warnings": []string{fmt.Sprintf("Failed to remove volumes: %v", volumeErr)},
				})
				return
			}
			utils.Result(fmt.Sprintf("Removed service '%s'", name))
			utils.Error(fmt.Sprintf("Failed to remove volumes: %v", volumeErr))
			return
		}
		utils.Result(fmt.Sprintf("Removed service '%s'", name))
	},
}

func init() {
	rootCmd.AddCommand(removeCmd)

	removeCmd.Flags().Bool("volumes", false, "Remove also volumes")
	addBulkFlags(removeCmd, "Remove")
}
