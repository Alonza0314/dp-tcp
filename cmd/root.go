package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "dp-tcp",
	Short: "Dual-Path TCP",
	Long:  "Dual-Path TCP is a layer 5 application for Dual-Path TCP.",
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
