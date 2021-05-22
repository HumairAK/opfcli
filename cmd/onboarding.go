package cmd

import (
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(onboardCmd)
}

var onboardCmd = &cobra.Command{
	Use:   "onboard",
	Short: "Onboard tools.",
}