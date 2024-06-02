package cmd

import (
	"bytes"
	_ "embed"
	"fmt"
	"github.com/stefanjarina/sda/internal/config"
	"github.com/stefanjarina/sda/internal/utils"
	"os"
	"path"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

//go:embed defaultConfig.yaml
var defaultCfgFile []byte

var rootCmd = &cobra.Command{
	Use:     "sda",
	Version: "0.0.4",
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
	rootCmd.PersistentFlags().Bool("json", false, "output as json")
	rootCmd.PersistentFlags().BoolP("yes", "y", false, "answer yes to all questions")
	_ = viper.BindPFlag("json", rootCmd.PersistentFlags().Lookup("json"))
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)
		cfgPath := path.Join(home, ".config", "sda")
		cfgFile = path.Join(cfgPath, "sda.yaml")

		if _, err := os.Stat(cfgPath); os.IsNotExist(err) {
			err := os.MkdirAll(cfgPath, 0755)
			if err != nil {
				utils.ErrorAndExit(fmt.Sprintf("Error creating config directory: %v", err))
			}
		}

		viper.SetConfigFile(cfgFile)
	}

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		saveConfig(defaultCfgFile)
	}

	if err := viper.Unmarshal(&config.CONFIG); err != nil {
		utils.ErrorAndExit(fmt.Sprintf("Error reading config file: %v", err))
	}
}

func saveConfig(defaultConfig []byte) {
	r := bytes.NewReader(defaultConfig)
	_ = viper.ReadConfig(r)

	if err := os.WriteFile(cfgFile, defaultConfig, 0644); err != nil {
		utils.ErrorAndExit(fmt.Sprintf("Error writing config file: %v", err))
	}
}
