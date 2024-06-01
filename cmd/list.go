package cmd

import (
	"fmt"
	"sda/internal/docker"
	"strings"

	"github.com/spf13/cobra"
)

type printFn func(docker.ServiceInfo)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List services (defaults to running)",
	Long:  `List services (defaults to running)`,
	Run: func(cmd *cobra.Command, args []string) {
		available, _ := cmd.Flags().GetBool("available")
		created, _ := cmd.Flags().GetBool("created")
		stopped, _ := cmd.Flags().GetBool("stopped")

		client := docker.New()

		if available {
			services := client.ListAvailable()
			printServices("Available services:", services, printListInfo)
			return
		}

		if created {
			services := client.ListCreated()
			printServices("Created services:", services, printServiceInfo)
			return
		}

		if stopped {
			services := client.ListStopped()
			printServices("Stopped services:", services, printServiceInfo)
			return
		}

		services := client.ListRunning()
		printServices("Running services:", services, printServiceInfo)
	},
}

func init() {
	rootCmd.AddCommand(listCmd)

	listCmd.Flags().BoolP("available", "a", false, "List available apps")
	listCmd.Flags().BoolP("created", "c", false, "List created apps")
	listCmd.Flags().BoolP("running", "r", false, "List running apps (default)")
	listCmd.Flags().BoolP("stopped", "s", false, "List stopped apps")
}

func printServices(title string, services []docker.ServiceInfo, fn printFn) {
	fmt.Println(title)

	for _, service := range services {
		fn(service)
	}
}

func printServiceInfo(service docker.ServiceInfo) {
	ports := strings.Join(service.Ports, ", ")
	fmt.Printf("%s\t%s\t\t(Image: '%s', Ports: [%s])\n", service.Name, service.Status, service.Image, ports)
}

func printListInfo(service docker.ServiceInfo) {
	fmt.Printf("%s - '%s:%s'\n", service.Name, service.Image, service.Version)
}
