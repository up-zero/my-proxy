/*
Package cmd
Copyright © 2025 getcharzp <getcharzp@gmail.com>
*/
package cmd

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
	"github.com/up-zero/my-proxy/models"
)

// createCmd represents the create command
var createCmd = &cobra.Command{
	Use:   "create <name>",
	Short: "create a new proxy service",
	Run: func(cmd *cobra.Command, args []string) {
		pb := &models.ProxyBasic{}
		if len(args) > 0 {
			pb.Name = args[0]
		}
		p := tea.NewProgram(initialModel(pb))
		if _, err := p.Run(); err != nil {
			fmt.Printf("TUI run error: %v\n", err)
			return
		}
	},
}

func init() {
	rootCmd.AddCommand(createCmd)
}
