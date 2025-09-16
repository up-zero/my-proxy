/*
Package cmd
Copyright © 2025 getcharzp <getcharzp@gmail.com>
*/
package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/up-zero/my-proxy/client/proxy"
	proxy2 "github.com/up-zero/my-proxy/service/proxy"
)

// startCmd represents the start command
var startCmd = &cobra.Command{
	Use:   "start <name>",
	Short: "start proxy service",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if err := proxy.Start(&proxy2.StartRequest{Name: args[0]}); err != nil {
			fmt.Printf("start error: %v\n", err)
			return
		}
		fmt.Println("start success")
	},
}

func init() {
	rootCmd.AddCommand(startCmd)
}
