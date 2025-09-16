/*
Package cmd
Copyright © 2025 getcharzp <getcharzp@gmail.com>
*/
package cmd

import (
	"fmt"
	"github.com/up-zero/my-proxy/util"

	"github.com/spf13/cobra"
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "get my-proxy version",
	Run: func(cmd *cobra.Command, args []string) {
		// my-proxy 1.0.0
		fmt.Println(fmt.Sprintf("%s %s", util.AppName, util.AppVersion))
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
