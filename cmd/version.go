package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	Version   = "0.1.0"
	RateCardV = "1.0"
	BuildDate = "dev"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version and rate card info",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("cctrack v%s\n", Version)
		fmt.Printf("Rate card v%s\n", RateCardV)
		fmt.Printf("Build date: %s\n", BuildDate)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
