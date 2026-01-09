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

		// Handle compose services
		if service != nil && service.IsComposeService() {
			build, _ := cmd.Flags().GetBool("build")
			recreate, _ := cmd.Flags().GetBool("recreate")
			removeVolumes, _ := cmd.Flags().GetBool("volumes")
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
						os.Exit(0)
					}
				}

				// Bring down the existing compose stack
				_ = client.ComposeDown(*service, removeVolumes)
			}

			if err := client.ComposeUp(*service, build, true); err != nil {
				utils.Error(fmt.Sprintf("Failed to create compose service '%s': %v", serviceName, err))
				utils.ErrorAndExit("")
			}
			fmt.Printf("Created and started service '%s'\n", serviceName)
			return
		}

		// Handle Docker services
		build, _ := cmd.Flags().GetBool("build")
		if build {
			fmt.Println("Warning: --build flag is only applicable for compose services, ignoring")
		}

		recreate, _ := cmd.Flags().GetBool("recreate")
		removeVolumes, _ := cmd.Flags().GetBool("volumes")
		yes, _ := cmd.Flags().GetBool("yes")

		if removeVolumes && !recreate {
			utils.Error("--volumes flag requires --recreate flag")
			utils.ErrorAndExit("")
		}

		// Get custom flags
		customPorts, _ := cmd.Flags().GetStringSlice("port")
		customVolumes, _ := cmd.Flags().GetStringSlice("volume")
		customEnvVars, _ := cmd.Flags().GetStringSlice("env")

		// Validate custom flags
		if len(customPorts) > 0 {
			portPattern := regexp.MustCompile(`^(\d+):(\d+)$|^(\d+\.\d+\.\d+\.\d+):(\d+):(\d+)$`)
			for _, port := range customPorts {
				if !portPattern.MatchString(port) {
					utils.Error(fmt.Sprintf("Invalid port format: %s (expected HOST:CONTAINER or IP:HOST:CONTAINER)", port))
					utils.ErrorAndExit("")
				}
			}
		}

		if len(customVolumes) > 0 {
			volumePattern := regexp.MustCompile(`^([^:]+):([^:]+)$`)
			for _, volume := range customVolumes {
				if !volumePattern.MatchString(volume) {
					utils.Error(fmt.Sprintf("Invalid volume format: %s (expected SOURCE:TARGET)", volume))
					utils.ErrorAndExit("")
				}
			}
		}

		if len(customEnvVars) > 0 {
			envPattern := regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*=.*$`)
			for _, env := range customEnvVars {
				if !envPattern.MatchString(env) {
					utils.Error(fmt.Sprintf("Invalid environment variable format: %s (expected KEY=VALUE)", env))
					utils.ErrorAndExit("")
				}
			}
		}

		cli := docker.New()

		if cli.Exists(serviceName) {
			if recreate {
				if removeVolumes {
					fmt.Printf("Recreating service '%s' and removing volumes...\n", serviceName)
				} else {
					fmt.Printf("Recreating service '%s' (volumes will be preserved)...\n", serviceName)
				}

				if !yes {
					confirmMsg := fmt.Sprintf("Recreate service '%s'? ", serviceName)
					if removeVolumes {
						confirmMsg += "This will remove all data. "
					}
					confirmMsg += "(Y/n): "

					if !utils.Confirm(confirmMsg) {
						os.Exit(0)
					}
				}

				fmt.Printf("Removing existing service '%s'...\n", serviceName)
				err := cli.Remove(serviceName)
				if err != nil {
					utils.Error(fmt.Sprintf("Failed to remove existing service: %v", err))
					utils.ErrorAndExit("")
				}

				if removeVolumes {
					service := config.CONFIG.GetServiceByName(serviceName)
					volumes := docker.GetNamedVolumesForService(service)
					if len(volumes) > 0 {
						fmt.Printf("Removing volumes: %s...\n", strings.Join(volumes, ", "))
						cli.RemoveVolumes(volumes)
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
	createCmd.Flags().Bool("recreate", false, "Remove existing container before creating")
	createCmd.Flags().Bool("volumes", false, "Also remove volumes when recreating (requires --recreate)")
	createCmd.Flags().Bool("build", false, "Build images before starting (compose services only)")
	createCmd.Flags().StringSlice("port", nil, "Port mapping (HOST:CONTAINER, can be specified multiple times)")
	createCmd.Flags().StringSlice("volume", nil, "Volume mapping (SOURCE:TARGET, can be specified multiple times)")
	createCmd.Flags().StringSliceP("env", "e", nil, "Environment variable (KEY=VALUE, can be specified multiple times)")
}
