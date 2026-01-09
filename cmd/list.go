package cmd

import (
	"fmt"

	"github.com/charmbracelet/bubbles/table"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
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
		stopped, _ := cmd.Flags().GetBool("stopped")
		noColor, _ := cmd.Flags().GetBool("no-color")
		format, _ := cmd.Flags().GetString("format")

		client := docker.New()

		var services []docker.ServiceInfo

		if available {
			services = client.ListAvailable()
		} else if created {
			services = client.ListCreated()
		} else if stopped {
			services = client.ListStopped()
		} else {
			services = client.ListRunning()
		}

		if viper.GetBool("json") || format == "json" {
			utils.Message(services)
			return
		}

		if len(services) == 0 {
			fmt.Println("No services found.")
			return
		}

		columns := []table.Column{
			{Title: "NAME", Width: 15},
			{Title: "STATUS", Width: 12},
			{Title: "VERSION", Width: 12},
			{Title: "CONTAINER", Width: 20},
			{Title: "PORTS", Width: 25},
		}

		rows := make([]table.Row, len(services))
		for i, s := range services {
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

			rows[i] = table.Row{
				s.Name,
				statusCell,
				s.Version,
				s.ContainerName,
				ports,
			}
		}

		t := table.New(
			table.WithColumns(columns),
			table.WithRows(rows),
			table.WithHeight(len(rows)),
		)

		fmt.Println(t.View())
	},
}

func init() {
	rootCmd.AddCommand(listCmd)

	listCmd.Flags().BoolP("available", "a", false, "List available apps")
	listCmd.Flags().BoolP("created", "c", false, "List created apps")
	listCmd.Flags().BoolP("running", "r", false, "List running apps (default)")
	listCmd.Flags().BoolP("stopped", "s", false, "List stopped apps")
	listCmd.Flags().Bool("no-color", false, "Disable color output")
	listCmd.Flags().String("format", "table", "Output format: table, json")
}
