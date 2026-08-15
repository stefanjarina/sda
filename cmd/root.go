package cmd

import (
	"bytes"
	_ "embed"
	"errors"
	"fmt"
	"os"
	"path"

	"github.com/stefanjarina/sda/internal/config"
	"github.com/stefanjarina/sda/internal/docker"
	"github.com/stefanjarina/sda/internal/utils"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var GitTag string
var cfgFile string

//go:embed defaultConfig.yaml
var defaultCfgFile []byte

var rootCmd = &cobra.Command{
	Use:     "sda",
	Version: GitTag,
	Short:   "Simple Docker Apps",
	Long:    `Simple Docker Apps`,
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

// GetRootCommand returns the root command for documentation generation
func GetRootCommand() *cobra.Command {
	return rootCmd
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.config/sda/sda.toml)")
	rootCmd.PersistentFlags().Bool("json", false, "output as json")
	rootCmd.PersistentFlags().BoolP("yes", "y", false, "answer yes to all questions")
	_ = viper.BindPFlag("json", rootCmd.PersistentFlags().Lookup("json"))
}

type configReadAction int

const (
	configWriteDefaults configReadAction = iota
	configFail
)

func classifyConfigReadError(err error, userSupplied bool) (configReadAction, string) {
	var notFound viper.ConfigFileNotFoundError
	missing := errors.As(err, &notFound) || errors.Is(err, os.ErrNotExist)
	if missing {
		if userSupplied {
			return configFail, fmt.Sprintf("Config file not found: %s", cfgFile)
		}
		return configWriteDefaults, ""
	}
	return configFail, fmt.Sprintf("Failed to read config %s: %v", cfgFile, err)
}

func initConfig() {
	userSupplied := cfgFile != ""
	if userSupplied {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)
		cfgPath := path.Join(home, ".config", "sda")
		cfgFile = path.Join(cfgPath, "sda.yaml")

		if _, err := os.Stat(cfgPath); os.IsNotExist(err) {
			if err := os.MkdirAll(cfgPath, 0755); err != nil {
				utils.ErrorAndExit(fmt.Sprintf("Error creating config directory: %v", err))
			}
		}

		viper.SetConfigFile(cfgFile)
	}

	viper.SetEnvPrefix("SDA")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		action, msg := classifyConfigReadError(err, userSupplied)
		switch action {
		case configWriteDefaults:
			saveConfig(defaultCfgFile)
		default:
			utils.ErrorAndExit(msg)
		}
	}

	if err := viper.Unmarshal(&config.CONFIG); err != nil {
		utils.ErrorAndExit(fmt.Sprintf("Error reading config file: %v", err))
	}

	for i := range config.CONFIG.Services {
		if err := docker.ValidateServiceTemplates(&config.CONFIG.Services[i]); err != nil {
			utils.ErrorAndExit(err.Error())
		}
	}
}

func saveConfig(defaultConfig []byte) {
	r := bytes.NewReader(defaultConfig)
	_ = viper.ReadConfig(r)

	if err := os.WriteFile(cfgFile, defaultConfig, 0644); err != nil {
		utils.ErrorAndExit(fmt.Sprintf("Error writing config file: %v", err))
	}
}
