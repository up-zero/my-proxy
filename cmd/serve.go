/*
Package cmd
Copyright © 2025 getcharzp <getcharzp@gmail.com>
*/
package cmd

import (
	"strings"

	"github.com/spf13/cobra"
	"github.com/up-zero/my-proxy/app"
	"github.com/up-zero/my-proxy/models"
)

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "run daemon process for proxy service",
	Run: func(cmd *cobra.Command, args []string) {
		servePort, _ := cmd.Flags().GetString("port")
		if servePort == "" {
			servePort = (&models.ConfigBasic{}).GetServerPort()
		}
		serveHost, _ := cmd.Flags().GetString("host")
		serveHost = strings.TrimSpace(serveHost)
		if serveHost == "" {
			serveHost = (&models.ConfigBasic{}).GetServerHost()
		}
		app.NewApp(serveHost, servePort)
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)
	serveCmd.Flags().StringP("port", "p", "", "service port")
	serveCmd.Flags().String("host", "", "service listen address, default 0.0.0.0 (all interfaces)")
}
