/*
Package cmd
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"github.com/up-zero/my-proxy/client/proxy"
	proxy2 "github.com/up-zero/my-proxy/service/proxy"

	"github.com/spf13/cobra"
)

// delCmd represents the del command
var delCmd = &cobra.Command{
	Use:   "del <name>",
	Short: "delete the existing proxy service",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if err := proxy.Delete(&proxy2.DeleteRequest{Name: args[0]}); err != nil {
			fmt.Printf("delete error: %v\n", err)
			return
		}
		fmt.Println("delete success")
	},
}

func init() {
	rootCmd.AddCommand(delCmd)
}
