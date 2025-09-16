/*
Package cmd
Copyright © 2025 getcharzp <getcharzp@gmail.com>
*/
package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "my-proxy",
	Short: "my-proxy CLI, support manage proxy, show proxy status etc.",
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
}
