package cmd

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/cnovak/carcharge/internal/services"
	"github.com/cnovak/carcharge/internal/util"
	"github.com/spf13/cobra"
)

// energyCmd represents the energy command
var energyCmd = &cobra.Command{
	Use:   "rebalance",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		var err error
		defer func() {
			if err != nil {
				log.Fatalln(err)
			}
		}()

		for {
			senseClient, err := services.NewSenseService(util.Config.Sense.Username, util.Config.Sense.Password)
			if err != nil {
				fmt.Fprintf(os.Stderr, "error: %v\n", err)
				os.Exit(1)
			}
			vehicleClient, err := services.GetVehicleClient()
			if err != nil {
				fmt.Fprintf(os.Stderr, "error: %v\n", err)
				os.Exit(1)
			}

			carService, err := services.NewTeslaService(vehicleClient)
			if err != nil {
				fmt.Fprintf(os.Stderr, "error: %v\n", err)
				os.Exit(1)
			}
			rebalancer, err := services.NewRebalancer(senseClient, carService)
			if err != nil {
				fmt.Fprintf(os.Stderr, "error: %v\n", err)
				os.Exit(1)
			}

			rebalancer.Rebalance()
			sleepSeconds := 30
			var plural string
			if sleepSeconds > 1 {
				plural = "s"
			} else {
				plural = ""
			}
			fmt.Printf("sleeping for %d second%s...", sleepSeconds, plural)
			time.Sleep(time.Duration(sleepSeconds) * time.Second)
		}
	},
}

func init() {
	rootCmd.AddCommand(energyCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// energyCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// energyCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
