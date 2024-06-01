/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"os"
	"sda/internal/config"
	"sda/internal/docker"

	"github.com/spf13/cobra"
)

// newCmd represents the new command
var newCmd = &cobra.Command{
	Use:   "new",
	Short: "Create new service",
	Long:  `Create new service`,
	Args:  cobra.ExactArgs(1),

	Run: func(cmd *cobra.Command, args []string) {
		serviceName := args[0]

		if !config.CONFIG.ServiceExists(serviceName) {
			fmt.Printf("Service '%s' does not exist\n", serviceName)
			os.Exit(1)
		}

		networkName, _ := cmd.Flags().GetString("network")
		password, _ := cmd.Flags().GetString("password")
		version, _ := cmd.Flags().GetString("version")

		if networkName != "" {
			config.CONFIG.UpdateNetwork(networkName)
		}

		if password != "" {
			config.CONFIG.UpdatePassword(password)
		}

		if version != "" {
			config.CONFIG.UpdateVersion(args[0], version)
		}

		cli := docker.New()

		if !cli.CheckNetwork() {
			cli.CreateNetwork()
		}
	},
}

func init() {
	rootCmd.AddCommand(newCmd)

	newCmd.Flags().StringP("network", "n", "", "Network name")
	newCmd.Flags().StringP("password", "p", "", "Password")
	newCmd.Flags().String("version", "", "Version")
}
