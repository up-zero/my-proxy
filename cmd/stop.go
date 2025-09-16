/*
Package cmd
Copyright © 2025 getcharzp <getcharzp@gmail.com>
*/
package cmd

import (
	"fmt"
	"github.com/up-zero/my-proxy/client/proxy"
	proxy2 "github.com/up-zero/my-proxy/service/proxy"

	"github.com/spf13/cobra"
)

// stopCmd represents the stop command
var stopCmd = &cobra.Command{
	Use:   "stop <name>",
	Short: "stop proxy service",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if err := proxy.Stop(&proxy2.StopRequest{Name: args[0]}); err != nil {
			fmt.Printf("stop error: %v\n", err)
			return
		}
		fmt.Println("stop success")
	},
}

func init() {
	rootCmd.AddCommand(stopCmd)
}
