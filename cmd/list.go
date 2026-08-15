package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/stefanjarina/sda/internal/config"
	"github.com/stefanjarina/sda/internal/docker"
	"github.com/stefanjarina/sda/internal/utils"
)

func cleanStatus(rawStatus string) string {
	if rawStatus == "" {
		return "available"
	}
	if len(rawStatus) >= 2 && (rawStatus[:2] == "Up" || rawStatus[:2] == "up") {
		return "running"
	}
	if len(rawStatus) >= 6 && rawStatus[:6] == "Exited" {
		return "stopped"
	}
	return rawStatus
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List services (defaults to running)",
	Long:  `List services (defaults to running)`,
	Run: func(cmd *cobra.Command, args []string) {
		available, _ := cmd.Flags().GetBool("available")
		created, _ := cmd.Flags().GetBool("created")
		running, _ := cmd.Flags().GetBool("running")
		stopped, _ := cmd.Flags().GetBool("stopped")
		compose, _ := cmd.Flags().GetBool("compose")
		noColor, _ := cmd.Flags().GetBool("no-color")

		mode, err := selectListMode(available, created, running, stopped, compose)
		if err != nil {
			utils.ErrorAndExit(err.Error())
		}

		client := docker.New()
		var services []docker.ServiceInfo

		switch mode {
		case listCompose:
			for _, svc := range config.CONFIG.Services {
				if svc.IsComposeService() {
					services = append(services, docker.ServiceInfo{
						Name:       svc.Name,
						Status:     "compose",
						Image:      "compose",
						Version:    "n/a",
						Ports:      []string{},
						StatusIcon: "📦",
					})
				}
			}
		case listAvailable:
			services = client.ListAvailable()
		default:
			var listErr error
			switch mode {
			case listCreated:
				services, listErr = client.ListCreated()
			case listStopped:
				services, listErr = client.ListStopped()
			default:
				services, listErr = client.ListRunning()
			}
			if listErr != nil {
				utils.ErrorAndExit(fmt.Sprintf("Failed to list services: %v", listErr))
			}
		}

		if wantsJSON(cmd) {
			utils.JSON(services)
			return
		}

		if len(services) == 0 {
			fmt.Println("No services found.")
			return
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)

		if available {
			fmt.Fprintln(w, "NAME\tVERSION\tIMAGE\t")
		} else {
			fmt.Fprintln(w, "NAME\tSTATUS\tVERSION\tCONTAINER\tPORTS\t")
		}

		for _, s := range services {
			ports := ""
			if len(s.Ports) > 0 {
				ports = s.Ports[0]
				for j := 1; j < len(s.Ports); j++ {
					ports += ", " + s.Ports[j]
				}
			}

			statusText := cleanStatus(s.Status)
			statusCell := s.StatusIcon + " " + statusText

			if !noColor {
				switch statusText {
				case "running":
					statusCell = "\033[32m" + statusCell + "\033[0m"
				case "stopped":
					statusCell = "\033[31m" + statusCell + "\033[0m"
				default:
					statusCell = "\033[33m" + statusCell + "\033[0m"
				}
			}

			if available {
				fmt.Fprintf(w, "%s\t%s\t%s\t\n",
					s.Name,
					s.Version,
					s.Image)
			} else {
				fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t\n",
					s.Name,
					statusCell,
					s.Version,
					s.ContainerName,
					ports)
			}
		}
		w.Flush()
	},
}

func init() {
	rootCmd.AddCommand(listCmd)

	listCmd.Flags().BoolP("available", "a", false, "List available apps")
	listCmd.Flags().BoolP("created", "c", false, "List created apps")
	listCmd.Flags().BoolP("running", "r", false, "List running apps (default)")
	listCmd.Flags().BoolP("stopped", "s", false, "List stopped apps")
	listCmd.Flags().Bool("compose", false, "List only compose services")
	listCmd.Flags().Bool("no-color", false, "Disable color output")
	listCmd.Flags().String("format", "table", "Output format: table, json")
}
