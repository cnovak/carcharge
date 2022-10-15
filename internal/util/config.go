/*
Copyright © 2022 Chris Novak <canovak@gmail.com>
*/
package util

import (
	"os"

	"github.com/apex/log"
	"github.com/apex/log/handlers/cli"
	"github.com/spf13/viper"
)

// Configuration
var Config Configuration

// Used for flags.
var ConfigFile string

type Configuration struct {
	Sense   SenseConfig
	Tesla   TeslaConfig
	Logfile string
}

type SenseConfig struct {
	Username string
	Password string
}

type TeslaConfig struct {
	AccessToken  string
	RefreshToken string
	ExpiresIn    int
	VehicleVin   string
	ChargerVolts int
}

func init() {
}

func InitializeConfig() {
	log.SetHandler(cli.Default)

	ctx := log.WithFields(log.Fields{
		"ConfigFile": ConfigFile,
	})

	if ConfigFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(ConfigFile)
		ctx.Info("use flag config file")
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		if err != nil {
			ctx.WithError(err).Error("cannot find home directory")
		}

		// Search config in home directory with name ".cobra" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName("config")
		viper.SetConfigType("yaml")
		viper.AddConfigPath(".")
	}

	viper.AutomaticEnv()

	err := viper.ReadInConfig()
	if err != nil { // Handle errors reading the config file
		ctx.WithError(err).Errorf("fatal error config file")
		os.Exit(1)
	}

	err = viper.Unmarshal(&Config)
	if err != nil {
		ctx.WithError(err).Error("Cannot unmarshal config file")
		os.Exit(1)
	}

	// TODO: Validate all fields are set
}
