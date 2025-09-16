/*
Package cmd
Copyright © 2025 getcharzp <getcharzp@gmail.com>
*/
package cmd

import (
	"github.com/spf13/cobra"
	"github.com/up-zero/my-proxy/app"
)

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "run daemon process for proxy service",
	Run: func(cmd *cobra.Command, args []string) {
		servePort, _ := cmd.Flags().GetString("port")
		app.NewApp(":" + servePort)
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)
	serveCmd.Flags().StringP("port", "p", "12312", "service port")
}
