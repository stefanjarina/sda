package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"sda/internal/docker"
	"sda/internal/utils"
)

// showCmd represents the show command
var showCmd = &cobra.Command{
	Use:   "show [service]",
	Short: "Show service information",
	Long:  `Show service information`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]
		client := docker.New()
		if client.Exists(name) {
			serviceInfo := client.GetInfo(name)

			if viper.GetBool("json") {
				utils.Message(serviceInfo)
				return
			} else {
				fmt.Printf("Name: %s\n", serviceInfo.Name)
				fmt.Printf("Status: %s\n", serviceInfo.Status)
				fmt.Printf("Image: %s\n", serviceInfo.Image)
				fmt.Printf("Ports: %v\n", serviceInfo.Ports)
				fmt.Printf("ID: %s\n", serviceInfo.ID)
				fmt.Printf("Container Name: %s\n", serviceInfo.ContainerName)
			}

		} else {
			fmt.Printf("Service %s not found\n", name)
		}
	},
}

func init() {
	rootCmd.AddCommand(showCmd)
}
