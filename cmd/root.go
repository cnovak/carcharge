package cmd

import (
	"github.com/apex/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	util "github.com/cnovak/carcharge/internal/util"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "carcharge",
	Short: "A brief description of your application",
	Long: `A longer description that spans multiple lines and likely contains
examples and usage of using your application. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		log.WithError(err).Errorf("Error executing command")

	}
}

func init() {

	rootCmd.PersistentFlags().StringVar(&util.ConfigFile, "config", "", "config file (default is ./config.yaml)")
	rootCmd.PersistentFlags().String("logfile", "./log.txt", "log file")

	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	viper.BindPFlag("logfile", rootCmd.PersistentFlags().Lookup("logfile"))

	cobra.OnInitialize(initialize)

}

func initialize() {
	util.InitializeConfig()
	util.InitializeLogs(util.Config)
}
