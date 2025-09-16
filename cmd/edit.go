/*
Package cmd
Copyright © 2025 getcharzp <getcharzp@gmail.com>
*/
package cmd

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/up-zero/my-proxy/client/proxy"

	"github.com/spf13/cobra"
)

// editCmd represents the edit command
var editCmd = &cobra.Command{
	Use:   "edit <name>",
	Short: "edit a existing proxy service",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// 获取代理详情
		pb, err := proxy.GetDetailByName(args[0])
		if err != nil {
			fmt.Printf("get proxy detail error: %v\n", err)
			return
		}
		p := tea.NewProgram(initialModel(pb))
		if _, err := p.Run(); err != nil {
			fmt.Printf("TUI run error: %v", err)
			return
		}
	},
}

func init() {
	rootCmd.AddCommand(editCmd)
}
