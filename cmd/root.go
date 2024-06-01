package cmd

import (
	"bytes"
	_ "embed"
	"fmt"
	"os"
	"path"
	"sda/internal/config"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

//go:embed defaultConfig.toml
var defaultCfgFile []byte

var rootCmd = &cobra.Command{
	Use:     "sda",
	Version: "0.0.1",
	Short:   "Simple Docker Apps",
	Long:    `Simple Docker Apps`,
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.config/sda/sda.toml)")
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)
		cfgPath := path.Join(home, ".config", "sda")
		cfgFile = path.Join(cfgPath, "sda.toml")

		if _, err := os.Stat(cfgPath); os.IsNotExist(err) {
			err := os.MkdirAll(cfgPath, 0755)
			if err != nil {
				fmt.Fprintln(os.Stderr, "Error creating config directory:", err)
				os.Exit(1)
			}
		}

		viper.SetConfigFile(cfgFile)
	}

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		saveConfig(defaultCfgFile)
	}

	if err := viper.Unmarshal(&config.CONFIG); err != nil {
		fmt.Fprintln(os.Stderr, "Error reading config file:", err)
	}
}

func saveConfig(defaultConfig []byte) {
	r := bytes.NewReader(defaultConfig)
	viper.ReadConfig(r)

	os.WriteFile(cfgFile, defaultConfig, 0644)

	if err := viper.WriteConfig(); err != nil {
		fmt.Fprintln(os.Stderr, "Error writing config file:", err)
		os.Exit(1)
	}
}
