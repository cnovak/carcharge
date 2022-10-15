/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/cnovak/carcharge/internal"
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
		for {
			senseClient, _ := internal.NewSenseService(util.Config.Sense.Username, util.Config.Sense.Password)
			vehicleClient, err := internal.GetVehicleClient()
			if err != nil {
				fmt.Fprintf(os.Stderr, "error: %v\n", err)
				os.Exit(1)
			}

			carService, err := internal.NewTeslaService(vehicleClient)
			if err != nil {
				fmt.Fprintf(os.Stderr, "error: %v\n", err)
				os.Exit(1)
			}
			rebalancer, err := internal.NewRebalancer(senseClient, carService)
			if err != nil {
				fmt.Fprintf(os.Stderr, "error: %v\n", err)
				os.Exit(1)
			}

			rebalancer.Rebalance()
			sleepMinutes := 2
			var plural string
			if sleepMinutes > 1 {
				plural = "s"
			} else {
				plural = ""
			}
			fmt.Printf("sleeping for %d minute%s...", sleepMinutes, plural)
			time.Sleep(time.Duration(sleepMinutes) * time.Minute)
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
