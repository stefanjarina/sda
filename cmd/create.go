package cmd

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

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

		service := config.CONFIG.GetServiceByName(serviceName)

		recreate, _ := cmd.Flags().GetBool("recreate")
		removeVolumes, _ := cmd.Flags().GetBool("volumes")
		if err := requireRecreateForVolumes(recreate, removeVolumes); err != nil {
			utils.ErrorAndExit(err.Error())
		}

		// Handle compose services
		if service != nil && service.IsComposeService() {
			build, _ := cmd.Flags().GetBool("build")
			yes, _ := cmd.Flags().GetBool("yes")

			client := docker.New()

			// If recreate is requested, we need to bring down the existing stack first
			if recreate {
				if !yes {
					confirmMsg := fmt.Sprintf("Recreate service '%s'? ", serviceName)
					if removeVolumes {
						confirmMsg += "This will remove all data. "
					}
					confirmMsg += "(Y/n): "

					if !utils.Confirm(confirmMsg) {
						utils.Cancelled()
					}
				}

				// Bring down the existing compose stack
				if err := client.ComposeDown(*service, removeVolumes); err != nil {
					utils.ErrorAndExit(fmt.Sprintf("Failed to bring down existing compose service '%s': %v", serviceName, err))
				}
			}

			if err := client.ComposeUp(*service, build, recreate); err != nil {
				utils.ErrorAndExit(fmt.Sprintf("Failed to create compose service '%s': %v", serviceName, err))
			}
			utils.Result(fmt.Sprintf("Created and started service '%s'", serviceName))
			return
		}

		// Handle Docker services
		build, _ := cmd.Flags().GetBool("build")
		if build {
			utils.Progress("Warning: --build flag is only applicable for compose services, ignoring\n")
		}

		yes, _ := cmd.Flags().GetBool("yes")

		// Get custom flags
		customPorts, _ := cmd.Flags().GetStringSlice("port")
		customVolumes, _ := cmd.Flags().GetStringSlice("volume")
		customEnvVars, _ := cmd.Flags().GetStringSlice("env")

		// Validate custom flags
		if len(customPorts) > 0 {
			portPattern := regexp.MustCompile(`^(\d+):(\d+)$|^(\d+\.\d+\.\d+\.\d+):(\d+):(\d+)$`)
			for _, port := range customPorts {
				if !portPattern.MatchString(port) {
					utils.ErrorAndExit(fmt.Sprintf("Invalid port format: %s (expected HOST:CONTAINER or IP:HOST:CONTAINER)", port))
				}
			}
		}

		if len(customVolumes) > 0 {
			volumePattern := regexp.MustCompile(`^([^:]+):([^:]+)$`)
			for _, volume := range customVolumes {
				if !volumePattern.MatchString(volume) {
					utils.ErrorAndExit(fmt.Sprintf("Invalid volume format: %s (expected SOURCE:TARGET)", volume))
				}
			}
		}

		if len(customEnvVars) > 0 {
			envPattern := regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*=.*$`)
			for _, env := range customEnvVars {
				if !envPattern.MatchString(env) {
					utils.ErrorAndExit(fmt.Sprintf("Invalid environment variable format: %s (expected KEY=VALUE)", env))
				}
			}
		}

		cli := docker.New()

		serviceExists, err := cli.Exists(serviceName)
		if err != nil {
			utils.ErrorAndExit(fmt.Sprintf("Failed to check if service '%s' exists: %v", serviceName, err))
		}

		if serviceExists {
			if recreate {
				if removeVolumes {
					utils.Progress("Recreating service '%s' and removing volumes...\n", serviceName)
				} else {
					utils.Progress("Recreating service '%s' (volumes will be preserved)...\n", serviceName)
				}

				if !yes {
					confirmMsg := fmt.Sprintf("Recreate service '%s'? ", serviceName)
					if removeVolumes {
						confirmMsg += "This will remove all data. "
					}
					confirmMsg += "(Y/n): "

					if !utils.Confirm(confirmMsg) {
						utils.Cancelled()
					}
				}

				utils.Progress("Removing existing service '%s'...\n", serviceName)
				err := cli.Remove(serviceName, removeVolumes)
				if err != nil {
					utils.ErrorAndExit(fmt.Sprintf("Failed to remove existing service: %v", err))
				}

				if removeVolumes {
					service := config.CONFIG.GetServiceByName(serviceName)
					volumes, err := docker.GetNamedVolumesForService(service)
					if err != nil {
						utils.ErrorAndExit(fmt.Sprintf("Failed to resolve volumes: %v", err))
					}
					if len(volumes) > 0 {
						utils.Progress("Removing volumes: %s...\n", strings.Join(volumes, ", "))
						if err := cli.RemoveVolumes(volumes); err != nil {
							if !utils.JSONMode() {
								utils.Error(fmt.Sprintf("Failed to remove volumes: %v", err))
							}
						}
					}
				}
			} else {
				utils.ErrorAndExit(fmt.Sprintf("Service '%s' already exists. Use --recreate to remove and recreate.", serviceName))
			}
		}

		// service is already declared earlier in this function
		service = config.CONFIG.GetServiceByName(serviceName)

		var networkName, password, version string
		var noStart bool
		networkName, _ = cmd.Flags().GetString("network")
		password, _ = cmd.Flags().GetString("password")
		version, _ = cmd.Flags().GetString("version")
		noStart, _ = cmd.Flags().GetBool("no-start")

		// Apply custom overrides to service config
		if len(customPorts) > 0 {
			var portMappings []config.PortMapping
			for _, port := range customPorts {
				parts := strings.Split(port, ":")
				if len(parts) == 2 {
					// HOST:CONTAINER format
					host, _ := strconv.Atoi(parts[0])
					container, _ := strconv.Atoi(parts[1])
					portMappings = append(portMappings, config.PortMapping{Host: host, Container: container})
				} else if len(parts) == 3 {
					// IP:HOST:CONTAINER format (ignore IP for now, just use HOST:CONTAINER)
					host, _ := strconv.Atoi(parts[1])
					container, _ := strconv.Atoi(parts[2])
					portMappings = append(portMappings, config.PortMapping{Host: host, Container: container})
				}
			}
			service.Docker.PortMappings = portMappings
		}

		if len(customVolumes) > 0 {
			var volumes []config.Volume
			for _, volume := range customVolumes {
				parts := strings.Split(volume, ":")
				source := parts[0]
				target := parts[1]
				isNamed := !strings.HasPrefix(source, "/") && !strings.HasPrefix(source, "\\") && !strings.Contains(source, ":\\")
				volumes = append(volumes, config.Volume{
					Source:  source,
					Target:  target,
					IsNamed: isNamed,
				})
			}
			service.Docker.Volumes = volumes
		}

		if len(customEnvVars) > 0 {
			service.Docker.EnvVars = customEnvVars
		}

		if service.HasPassword {
			if password != "" {
				config.CONFIG.UpdatePassword(password)
				utils.Progress("Creating '%s' with custom password\n", service.OutputName)
			} else {
				utils.Progress("Creating '%s' with default password\n", service.OutputName)
				utils.Progress("For custom password run: 'sda create %s -p <PASSWORD>'\n", serviceName)
				utils.Progress("Password must be strong, otherwise Docker fails to create container\n")
			}
		} else {
			utils.Progress("Creating '%s'\n", service.OutputName)
		}

		if !yes {
			answer := utils.Confirm("Proceed? (Y/n): ")
			if !answer {
				utils.Cancelled()
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
			if err := cli.CreateNetwork(); err != nil {
				utils.ErrorAndExit(fmt.Sprintf("Failed to create network '%s': %v", config.CONFIG.Network, err))
			}
		}

		if err := cli.Create(serviceName); err != nil {
			utils.ErrorAndExit(fmt.Sprintf("Failed to create container: %v", err))
		}

		if noStart {
			utils.Result(fmt.Sprintf("Created service '%s'", service.OutputName))
			os.Exit(0)
		}
		utils.Progress("Created service '%s'\n", service.OutputName)

		if err := cli.Start(serviceName); err != nil {
			utils.ErrorAndExit(fmt.Sprintf("Failed to start container: %v", err))
		}

		utils.Result(fmt.Sprintf("Started service '%s'", service.OutputName))
	},
}

func init() {
	rootCmd.AddCommand(createCmd)

	createCmd.Flags().StringP("network", "n", "", "Network name")
	createCmd.Flags().StringP("password", "p", "", "Password")
	createCmd.Flags().String("version", "", "Version")
	createCmd.Flags().Bool("no-start", false, "Do not start container after creation")
	createCmd.Flags().Bool("recreate", false, "Remove existing container before creating")
	createCmd.Flags().Bool("volumes", false, "Also remove volumes when recreating (requires --recreate)")
	createCmd.Flags().Bool("build", false, "Build images before starting (compose services only)")
	createCmd.Flags().StringSlice("port", nil, "Port mapping (HOST:CONTAINER, can be specified multiple times)")
	createCmd.Flags().StringSlice("volume", nil, "Volume mapping (SOURCE:TARGET, can be specified multiple times)")
	createCmd.Flags().StringSliceP("env", "e", nil, "Environment variable (KEY=VALUE, can be specified multiple times)")
}
