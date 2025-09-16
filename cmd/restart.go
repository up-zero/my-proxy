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

// restartCmd represents the restart command
var restartCmd = &cobra.Command{
	Use:   "restart <name>",
	Short: "restart proxy service",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if err := proxy.Restart(&proxy2.RestartRequest{Name: args[0]}); err != nil {
			fmt.Printf("restart error: %v\n", err)
			return
		}
		fmt.Println("restart success")
	},
}

func init() {
	rootCmd.AddCommand(restartCmd)
}
