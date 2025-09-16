/*
Package cmd
Copyright © 2025 getcharzp <getcharzp@gmail.com>
*/
package cmd

import (
	"fmt"
	"github.com/olekukonko/tablewriter"
	"github.com/up-zero/my-proxy/client/proxy"
	"os"
	"strconv"

	"github.com/spf13/cobra"
)

// statusCmd represents the status command
var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "get proxy service status, default get all service status",
	Run: func(cmd *cobra.Command, args []string) {
		var name string
		if len(args) > 0 {
			name = args[0]
		}
		tasks, err := proxy.Status(name)
		if err != nil {
			fmt.Printf("get status error: %v\n", err)
			return
		}

		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"#", "Name", "Type", "Status", "Listen Port", "Target Address", "Target Port"})
		for idx, task := range tasks {
			table.Append([]string{
				strconv.Itoa(idx + 1),
				task.Name,
				task.Type,
				task.State,
				task.ListenPort,
				task.TargetAddress,
				task.TargetPort,
			})
		}
		table.Render()
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
}
