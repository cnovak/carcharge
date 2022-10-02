package myutil

import (
	"fmt"
	"log"
	"os"

	"github.com/spf13/viper"
)

// global variables available via util package
var (
	Port  int
	DbURI string
)

var CfgFile string

// initConfig reads in config file and ENV variables if set.
func InitConfig() {
	if CfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(CfgFile)
	} else {
		// Find home directory.
		// home, err := os.UserHomeDir()
		// cobra.CheckErr(err)

		// Search config in home directory with name ".carcharge" (without extension).
		viper.SetConfigFile(".env")
		// viper.AddConfigPath(home)
		// viper.SetConfigType("yaml")
		// viper.SetConfigName(".carcharge")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
	fmt.Println("API TOKEN:", viper.Get("API_SECRET"))
}

func init() {
	// viper.SetDefault(PORT, 8080)
	viper.SetConfigFile(".env")
	viper.AddConfigPath(".")
	viper.AutomaticEnv()

	fmt.Println("---------to see in test printout")
	cwd, _ := os.Getwd()
	fmt.Println(cwd)
	fmt.Println("---------")

	if err := viper.ReadInConfig(); err != nil {
		log.Fatal("no environment file!")
	}

	Port = viper.GetInt("PORT")
	DbURI = viper.GetString("DB_URI")
}
