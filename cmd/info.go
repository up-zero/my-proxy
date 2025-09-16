/*
Package cmd
Copyright © 2025 getcharzp <getcharzp@gmail.com>
*/
package cmd

import (
	"fmt"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"github.com/up-zero/my-proxy/client/info"
	"github.com/up-zero/my-proxy/util"
	"os"
)

// infoCmd represents the info command
var infoCmd = &cobra.Command{
	Use:   "info",
	Short: "get detailed information",
	Run: func(cmd *cobra.Command, args []string) {
		// my-proxy 1.0.0
		fmt.Println(fmt.Sprintf("%s %s", util.AppName, util.AppVersion))
		reply, err := info.Info()
		if err != nil {
			fmt.Printf("get info error: %v \n", err)
			return
		}

		table := tablewriter.NewWriter(os.Stdout)
		table.SetAutoMergeCells(true)
		for _, address := range reply.Addresses {
			table.Append([]string{"Address", address})
		}
		table.Append([]string{"Username", reply.Username})
		table.Append([]string{"Password", reply.Password})

		table.Render()
	},
}

func init() {
	rootCmd.AddCommand(infoCmd)
}
